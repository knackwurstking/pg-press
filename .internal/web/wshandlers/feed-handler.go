package wshandlers

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers/templates"
	"github.com/knackwurstking/pgpress/pkg/logger"

	"golang.org/x/net/websocket"
)

// FeedCounterTemplateData represents the data for rendering feed counter template
type FeedCounterTemplateData struct {
	Count int
}

// FeedHandler manages WebSocket connections for feed counter updates
type FeedHandler struct {
	connections map[*FeedConnection]bool
	register    chan *FeedConnection
	unregister  chan *FeedConnection
	broadcast   chan struct{}
	db          *services.Registry
	mu          sync.RWMutex

	log *logger.Logger
}

// NewFeedHandler creates a new feed notification manager
func NewFeedHandler(db *services.Registry) *FeedHandler {
	return &FeedHandler{
		connections: make(map[*FeedConnection]bool),
		register:    make(chan *FeedConnection),
		unregister:  make(chan *FeedConnection),
		broadcast:   make(chan struct{}, 100), // Buffered channel to prevent blocking
		db:          db,
		log:         logger.GetComponentLogger("WS Feeds"),
	}
}

// Start begins the notification manager's main loop
func (fn *FeedHandler) Start(ctx context.Context) {
	fn.log.Info("Starting feed notification manager")

	for {
		select {
		case <-ctx.Done():
			fn.log.Info("Shutting down feed notification manager")
			fn.closeAllConnections()
			return
		case conn := <-fn.register:
			fn.mu.Lock()
			fn.connections[conn] = true
			fn.mu.Unlock()
			fn.log.Info("Registered new connection for user ID %d", conn.UserID)

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
			fn.log.Info("Unregistered connection for user ID %d", conn.UserID)

		case <-fn.broadcast:
			fn.broadcastToAllConnections()
		}
	}
}

// RegisterConnection adds a new WebSocket connection to the manager
func (fn *FeedHandler) RegisterConnection(
	userID, lastFeed int64,
	conn *websocket.Conn,
) *FeedConnection {
	feedConn := NewFeedConnection(userID, conn)
	feedConn.LastFeed = lastFeed

	fn.register <- feedConn
	return feedConn
}

// UnregisterConnection removes a WebSocket connection from the manager
func (fn *FeedHandler) UnregisterConnection(conn *FeedConnection) {
	fn.unregister <- conn
}

// Broadcast broadcasts feed counter updates to all connected clients
func (fn *FeedHandler) Broadcast() {
	select {
	case fn.broadcast <- struct{}{}:
		fn.log.Info("Feed update notification queued")
	default:
		fn.log.Warn("Broadcast channel full, skipping notification")
	}
}

// sendInitialFeedCounter sends the initial feed counter to a newly connected client
func (fn *FeedHandler) sendInitialFeedCounter(conn *FeedConnection) {
	html, err := fn.renderFeedCounter(conn.LastFeed)
	if err != nil {
		fn.log.Error("Error rendering initial feed counter for user %d: %v", conn.UserID, err)
		return
	}

	select {
	case conn.Send <- html:
		// Successfully queued message
	case <-conn.Done:
		// Connection was closed
		return
	case <-time.After(30 * time.Second):
		fn.log.Warn("Timeout sending initial feed counter to user %d", conn.UserID)
	}
}

// broadcastToAllConnections sends feed counter updates to all connected clients
func (fn *FeedHandler) broadcastToAllConnections() {
	fn.mu.RLock()
	connections := make([]*FeedConnection, 0, len(fn.connections))
	for conn := range fn.connections {
		connections = append(connections, conn)
	}
	fn.mu.RUnlock()

	fn.log.Info("Broadcasting feed counter update to %d connections", len(connections))

	for _, conn := range connections {
		go fn.sendFeedCounterUpdate(conn)
	}
}

// sendFeedCounterUpdate sends an updated feed counter to a specific connection
func (fn *FeedHandler) sendFeedCounterUpdate(conn *FeedConnection) {
	html, err := fn.renderFeedCounter(conn.LastFeed)
	if err != nil {
		fn.log.Error("Error rendering feed counter for user %d: %v", conn.UserID, err)
		return
	}

	select {
	case conn.Send <- html:
		// Successfully queued message
	case <-conn.Done:
		// Connection was closed, unregister it
		fn.UnregisterConnection(conn)
	case <-time.After(30 * time.Second):
		fn.log.Warn("Timeout sending feed counter update to user %d", conn.UserID)
		// Don't unregister on timeout, the connection might just be slow or suspended
	}
}

// renderFeedCounter renders the feed counter template with current data
func (fn *FeedHandler) renderFeedCounter(userLastFeed int64) ([]byte, error) {
	feeds, err := fn.db.Feeds.ListRange(0, constants.MaxFeedsPerPage)
	if err != nil {
		return nil, err
	}

	count := int(0)
	for _, feed := range feeds {
		if feed.ID <= userLastFeed {
			break
		}
		count++
	}

	var buf bytes.Buffer
	err = templates.FeedCounter(count).Render(context.Background(), &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// closeAllConnections gracefully closes all WebSocket connections
func (fn *FeedHandler) closeAllConnections() {
	fn.mu.Lock()
	defer fn.mu.Unlock()

	for conn := range fn.connections {
		conn.Conn.Close()
		close(conn.Send)
		close(conn.Done)
		delete(fn.connections, conn)
	}

	fn.log.Info("Closed all WebSocket connections")
}

// ConnectionCount returns the number of active connections
func (fn *FeedHandler) ConnectionCount() int {
	fn.mu.RLock()
	defer fn.mu.RUnlock()
	return len(fn.connections)
}

// FeedConnection represents a WebSocket connection for feed updates
type FeedConnection struct {
	UserID   int64
	LastFeed int64
	Conn     *websocket.Conn
	Send     chan []byte
	Done     chan struct{}

	log *logger.Logger
}

func NewFeedConnection(userID int64, conn *websocket.Conn) *FeedConnection {
	return &FeedConnection{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Done:   make(chan struct{}),
		log:    logger.GetComponentLogger("WS Connection"),
	}
}

// WritePump handles writing messages to the WebSocket connection
func (conn *FeedConnection) WritePump() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer conn.Conn.Close()

	for {
		select {
		case message, ok := <-conn.Send:
			if !ok {
				// The channel was closed
				return
			}

			// Set write deadline - longer timeout for suspended connections
			if err := conn.Conn.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
				conn.log.Error("Error setting write deadline for user %d: %v", conn.UserID, err)
				return
			}

			if err := websocket.Message.Send(conn.Conn, string(message)); err != nil {
				conn.log.Error("Error writing message to user %d: %v", conn.UserID, err)
				return
			}

		case <-ticker.C:
			// Send ping message - golang.org/x/net/websocket doesn't have built-in ping/pong
			// We'll send a simple ping message instead
			if err := conn.Conn.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
				conn.log.Error("Error setting ping write deadline for user %d: %v", conn.UserID, err)
				return
			}
			if err := websocket.Message.Send(conn.Conn, "ping"); err != nil {
				conn.log.Error("Error sending ping to user %d: %v", conn.UserID, err)
				return
			}

		case <-conn.Done:
			return
		}
	}
}

// ReadPump handles reading messages from the WebSocket connection
func (conn *FeedConnection) ReadPump(handler *FeedHandler) {
	defer func() {
		handler.UnregisterConnection(conn)
		conn.Conn.Close()
	}()

	for {
		var message string
		err := websocket.Message.Receive(conn.Conn, &message)
		if err != nil {
			if err == io.EOF {
				conn.log.Info("Connection closed normally for user %d", conn.UserID)
			} else {
				// Check if error is due to timeout or suspension-related issues
				if isTemporaryError(err) {
					conn.log.Warn(
						"Temporary error for user %d (possibly suspended): %v",
						conn.UserID, err)
					// Continue without breaking - browser might be suspended
					continue
				}
				conn.log.Error("Error reading message from user %d: %v", conn.UserID, err)
			}
			break
		}

		// Handle ping/pong manually since golang.org/x/net/websocket doesn't have built-in support
		if message == "ping" {
			// Respond with pong
			if err := websocket.Message.Send(conn.Conn, "pong"); err != nil {
				conn.log.Error("Error sending pong to user %d: %v", conn.UserID, err)
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

// isTemporaryError checks if an error is temporary and the connection should be kept alive
func isTemporaryError(err error) bool {
	// Check for common suspension-related errors
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset")
}
