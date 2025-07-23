package notifications

import (
	"context"
	"io"
	"io/fs"
	"log"
	"sync"
	"time"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"golang.org/x/net/websocket"
)

// FeedConnection represents a WebSocket connection for feed updates
type FeedConnection struct {
	UserID   int64
	LastFeed int64
	Conn     *websocket.Conn
	Send     chan []byte
	Done     chan struct{}
}

// FeedNotifier manages WebSocket connections for feed counter updates
type FeedNotifier struct {
	connections map[*FeedConnection]bool
	register    chan *FeedConnection
	unregister  chan *FeedConnection
	broadcast   chan struct{}
	db          *pgvis.DB
	templates   fs.FS
	mu          sync.RWMutex
}

// FeedCounterTemplateData represents the data for rendering feed counter template
type FeedCounterTemplateData struct {
	Count int
}

// NewFeedNotifier creates a new feed notification manager
func NewFeedNotifier(db *pgvis.DB, templates fs.FS) *FeedNotifier {
	return &FeedNotifier{
		connections: make(map[*FeedConnection]bool),
		register:    make(chan *FeedConnection),
		unregister:  make(chan *FeedConnection),
		broadcast:   make(chan struct{}, 100), // Buffered channel to prevent blocking
		db:          db,
		templates:   templates,
	}
}

// Start begins the notification manager's main loop
func (fn *FeedNotifier) Start(ctx context.Context) {
	log.Printf("[FeedNotifier] Starting feed notification manager")

	for {
		select {
		case <-ctx.Done():
			log.Printf("[FeedNotifier] Shutting down feed notification manager")
			fn.closeAllConnections()
			return
		case conn := <-fn.register:
			fn.mu.Lock()
			fn.connections[conn] = true
			fn.mu.Unlock()
			log.Printf("[FeedNotifier] Registered new connection for user ID %d", conn.UserID)

			// Send initial feed counter to new connection
			go fn.sendInitialFeedCounter(conn)

		case conn := <-fn.unregister:
			fn.mu.Lock()
			if _, ok := fn.connections[conn]; ok {
				delete(fn.connections, conn)
				close(conn.Send)
				close(conn.Done)
			}
			fn.mu.Unlock()
			log.Printf("[FeedNotifier] Unregistered connection for user ID %d", conn.UserID)

		case <-fn.broadcast:
			fn.broadcastToAllConnections()
		}
	}
}

// RegisterConnection adds a new WebSocket connection to the manager
func (fn *FeedNotifier) RegisterConnection(userID, lastFeed int64, conn *websocket.Conn) *FeedConnection {
	feedConn := &FeedConnection{
		UserID:   userID,
		LastFeed: lastFeed,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Done:     make(chan struct{}),
	}

	fn.register <- feedConn
	return feedConn
}

// UnregisterConnection removes a WebSocket connection from the manager
func (fn *FeedNotifier) UnregisterConnection(conn *FeedConnection) {
	fn.unregister <- conn
}

// NotifyNewFeed broadcasts feed counter updates to all connected clients
func (fn *FeedNotifier) NotifyNewFeed() {
	select {
	case fn.broadcast <- struct{}{}:
		log.Printf("[FeedNotifier] Feed update notification queued")
	default:
		log.Printf("[FeedNotifier] Broadcast channel full, skipping notification")
	}
}

// sendInitialFeedCounter sends the initial feed counter to a newly connected client
func (fn *FeedNotifier) sendInitialFeedCounter(conn *FeedConnection) {
	html, err := fn.renderFeedCounter(conn.LastFeed)
	if err != nil {
		log.Printf("[FeedNotifier] Error rendering initial feed counter for user %d: %v", conn.UserID, err)
		return
	}

	select {
	case conn.Send <- html:
		// Successfully queued message
	case <-conn.Done:
		// Connection was closed
		return
	case <-time.After(5 * time.Second):
		log.Printf("[FeedNotifier] Timeout sending initial feed counter to user %d", conn.UserID)
	}
}

// broadcastToAllConnections sends feed counter updates to all connected clients
func (fn *FeedNotifier) broadcastToAllConnections() {
	fn.mu.RLock()
	connections := make([]*FeedConnection, 0, len(fn.connections))
	for conn := range fn.connections {
		connections = append(connections, conn)
	}
	fn.mu.RUnlock()

	log.Printf("[FeedNotifier] Broadcasting feed counter update to %d connections", len(connections))

	for _, conn := range connections {
		go fn.sendFeedCounterUpdate(conn)
	}
}

// sendFeedCounterUpdate sends an updated feed counter to a specific connection
func (fn *FeedNotifier) sendFeedCounterUpdate(conn *FeedConnection) {
	html, err := fn.renderFeedCounter(conn.LastFeed)
	if err != nil {
		log.Printf("[FeedNotifier] Error rendering feed counter for user %d: %v", conn.UserID, err)
		return
	}

	select {
	case conn.Send <- html:
		// Successfully queued message
	case <-conn.Done:
		// Connection was closed, unregister it
		fn.UnregisterConnection(conn)
	case <-time.After(5 * time.Second):
		log.Printf("[FeedNotifier] Timeout sending feed counter update to user %d", conn.UserID)
		// Don't unregister on timeout, the connection might just be slow
	}
}

// renderFeedCounter renders the feed counter template with current data
func (fn *FeedNotifier) renderFeedCounter(userLastFeed int64) ([]byte, error) {
	data := &FeedCounterTemplateData{}

	feeds, err := fn.db.Feeds.ListRange(0, 100)
	if err != nil {
		return nil, err
	}

	for _, feed := range feeds {
		if feed.ID > userLastFeed {
			data.Count++
		} else {
			break
		}
	}

	html, err := utils.RenderTemplateToString(
		fn.templates,
		[]string{constants.FeedCounterComponentTemplatePath},
		data,
	)
	if err != nil {
		return nil, err
	}

	return []byte(html), nil
}

// closeAllConnections gracefully closes all WebSocket connections
func (fn *FeedNotifier) closeAllConnections() {
	fn.mu.Lock()
	defer fn.mu.Unlock()

	for conn := range fn.connections {
		conn.Conn.Close()
		close(conn.Send)
		close(conn.Done)
		delete(fn.connections, conn)
	}

	log.Printf("[FeedNotifier] Closed all WebSocket connections")
}

// ConnectionCount returns the number of active connections
func (fn *FeedNotifier) ConnectionCount() int {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	return len(fn.connections)
}

// WritePump handles writing messages to the WebSocket connection
func (conn *FeedConnection) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()
	defer conn.Conn.Close()

	for {
		select {
		case message, ok := <-conn.Send:
			if !ok {
				// The channel was closed
				return
			}

			// Set write deadline
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if err := websocket.Message.Send(conn.Conn, string(message)); err != nil {
				log.Printf("[FeedConnection] Error writing message to user %d: %v", conn.UserID, err)
				return
			}

		case <-ticker.C:
			// Send ping message - golang.org/x/net/websocket doesn't have built-in ping/pong
			// We'll send a simple ping message instead
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := websocket.Message.Send(conn.Conn, "ping"); err != nil {
				log.Printf("[FeedConnection] Error sending ping to user %d: %v", conn.UserID, err)
				return
			}

		case <-conn.Done:
			return
		}
	}
}

// ReadPump handles reading messages from the WebSocket connection
func (conn *FeedConnection) ReadPump(notifier *FeedNotifier) {
	defer func() {
		notifier.UnregisterConnection(conn)
		conn.Conn.Close()
	}()

	// Set read deadline
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		var message string
		err := websocket.Message.Receive(conn.Conn, &message)
		if err != nil {
			if err == io.EOF {
				log.Printf("[FeedConnection] Connection closed normally for user %d", conn.UserID)
			} else {
				log.Printf("[FeedConnection] Error reading message from user %d: %v", conn.UserID, err)
			}
			break
		}

		// Reset read deadline on successful message
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Handle ping/pong manually since golang.org/x/net/websocket doesn't have built-in support
		if message == "ping" {
			// Respond with pong
			if err := websocket.Message.Send(conn.Conn, "pong"); err != nil {
				log.Printf("[FeedConnection] Error sending pong to user %d: %v", conn.UserID, err)
				break
			}
		} else if message == "pong" {
			// Received pong response, connection is alive
			continue
		}

		// Client sent a message, could trigger immediate update if needed
		// For now, we just acknowledge that the connection is active
	}
}

// CreateWebSocketHandler creates a WebSocket handler function for use with net/http
func (fn *FeedNotifier) CreateWebSocketHandler(getUserFromRequest func(ws *websocket.Conn) (*pgvis.User, error)) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		// Get user from request context
		user, err := getUserFromRequest(ws)
		if err != nil {
			log.Printf("[WebSocket] User authentication failed: %v", err)
			ws.Close()
			return
		}

		log.Printf("[WebSocket] User authenticated: %s (LastFeed: %d)", user.UserName, user.LastFeed)

		// Register the connection with the feed notifier
		feedConn := fn.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			log.Printf("[WebSocket] Failed to register connection for user %s", user.UserName)
			ws.Close()
			return
		}

		log.Printf("[WebSocket] Connection registered for user %s", user.UserName)

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(fn)

		log.Printf("[WebSocket] Connection closed for user %s", user.UserName)
	})
}
