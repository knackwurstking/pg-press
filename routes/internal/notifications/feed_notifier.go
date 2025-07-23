package notifications

import (
	"context"
	"io/fs"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
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
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The channel was closed
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[FeedConnection] Error writing message to user %d: %v", conn.UserID, err)
				return
			}

		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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

	conn.Conn.SetReadLimit(512)
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[FeedConnection] Unexpected close error for user %d: %v", conn.UserID, err)
			}
			break
		}
		// Client sent a message, could trigger immediate update if needed
	}
}
