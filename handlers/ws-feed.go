package handlers

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/services"

	"golang.org/x/net/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeTimeout = 30 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// FeedHandler manages WebSocket connections for feed counter updates
type FeedHandler struct {
	connections map[*FeedConnection]bool
	register    chan *FeedConnection
	unregister  chan *FeedConnection
	broadcast   chan struct{}
	db          *services.Registry
	mu          sync.RWMutex
	log         *logger.Logger
}

// NewFeedHandler creates a new feed notification manager
func NewFeedHandler(db *services.Registry) *FeedHandler {
	return &FeedHandler{
		connections: make(map[*FeedConnection]bool),
		register:    make(chan *FeedConnection),
		unregister:  make(chan *FeedConnection),
		broadcast:   make(chan struct{}, 100),
		db:          db,
		log:         logger.NewComponentLogger("WS: Feed"),
	}
}

// Start begins the notification manager's main loop
func (fh *FeedHandler) Start(ctx context.Context) {
	fh.log.Info("Starting feed notification manager")

	for {
		select {
		case <-ctx.Done():
			fh.log.Info("Shutting down feed notification manager")
			fh.closeAllConnections()
			return

		case conn := <-fh.register:
			fh.registerConnection(conn)

		case conn := <-fh.unregister:
			fh.unregisterConnection(conn)

		case <-fh.broadcast:
			fh.broadcastToAllConnections()
		}
	}
}

// RegisterConnection adds a new WebSocket connection to the manager
func (fh *FeedHandler) RegisterConnection(userID, lastFeed int64, conn *websocket.Conn) *FeedConnection {
	feedConn := NewFeedConnection(userID, lastFeed, conn)
	fh.register <- feedConn
	return feedConn
}

// UnregisterConnection removes a WebSocket connection from the manager
func (fh *FeedHandler) UnregisterConnection(conn *FeedConnection) {
	fh.unregister <- conn
}

// Broadcast queues a feed counter update for all connected clients
func (fh *FeedHandler) Broadcast() {
	select {
	case fh.broadcast <- struct{}{}:
		fh.log.Debug("Feed update notification queued")
	default:
		fh.log.Warn("Broadcast channel full, skipping notification")
	}
}

// ConnectionCount returns the number of active connections
func (fh *FeedHandler) ConnectionCount() int {
	fh.mu.RLock()
	defer fh.mu.RUnlock()
	return len(fh.connections)
}

func (fh *FeedHandler) registerConnection(conn *FeedConnection) {
	fh.mu.Lock()
	fh.connections[conn] = true
	fh.mu.Unlock()

	fh.log.Info("Registered new connection for user ID %d", conn.UserID)
	go fh.sendUpdate(conn)
}

func (fh *FeedHandler) unregisterConnection(conn *FeedConnection) {
	fh.mu.Lock()
	if _, ok := fh.connections[conn]; ok {
		delete(fh.connections, conn)
		close(conn.send)
		close(conn.done)
	}
	fh.mu.Unlock()

	fh.log.Info("Unregistered connection for user ID %d", conn.UserID)
}

func (fh *FeedHandler) broadcastToAllConnections() {
	fh.mu.RLock()
	connections := make([]*FeedConnection, 0, len(fh.connections))
	for conn := range fh.connections {
		connections = append(connections, conn)
	}
	fh.mu.RUnlock()

	fh.log.Debug("Broadcasting feed counter update to %d connections", len(connections))

	for _, conn := range connections {
		go fh.sendUpdate(conn)
	}
}

func (fh *FeedHandler) sendUpdate(conn *FeedConnection) {
	html, err := fh.renderFeedCounter(conn.LastFeed)
	if err != nil {
		fh.log.Error("Error rendering feed counter for user %d: %v", conn.UserID, err)
		return
	}

	select {
	case conn.send <- html:
		// Successfully queued
	case <-conn.done:
		// Connection closed
		fh.UnregisterConnection(conn)
	case <-time.After(writeTimeout):
		fh.log.Warn("Timeout sending feed counter to user %d", conn.UserID)
	}
}

func (fh *FeedHandler) renderFeedCounter(userLastFeed int64) ([]byte, error) {
	feeds, err := fh.db.Feeds.ListRange(0, env.MaxFeedsPerPage)
	if err != nil {
		return nil, err
	}

	count := 0
	for _, feed := range feeds {
		if feed.ID <= userLastFeed {
			break
		}
		count++
	}

	var buf bytes.Buffer
	if err := components.FeedCounter(count).Render(context.Background(), &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (fh *FeedHandler) closeAllConnections() {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	for conn := range fh.connections {
		conn.conn.Close()
		close(conn.send)
		close(conn.done)
		delete(fh.connections, conn)
	}

	fh.log.Info("Closed all WebSocket connections")
}

// FeedConnection represents a WebSocket connection for feed updates
type FeedConnection struct {
	UserID   int64
	LastFeed int64
	conn     *websocket.Conn
	send     chan []byte
	done     chan struct{}
	log      *logger.Logger
}

// NewFeedConnection creates a new feed connection
func NewFeedConnection(userID, lastFeed int64, conn *websocket.Conn) *FeedConnection {
	return &FeedConnection{
		UserID:   userID,
		LastFeed: lastFeed,
		conn:     conn,
		send:     make(chan []byte, 256),
		done:     make(chan struct{}),
		log:      logger.NewComponentLogger("WS: Connection"),
	}
}

// WritePump handles writing messages to the WebSocket connection
func (fc *FeedConnection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		fc.conn.Close()
	}()

	for {
		select {
		case message, ok := <-fc.send:
			if !ok {
				return
			}

			if err := fc.writeMessage(string(message)); err != nil {
				fc.log.Error("Error writing message to user %d: %v", fc.UserID, err)
				return
			}

		case <-ticker.C:
			if err := fc.writeMessage("ping"); err != nil {
				fc.log.Error("Error sending ping to user %d: %v", fc.UserID, err)
				return
			}

		case <-fc.done:
			return
		}
	}
}

// ReadPump handles reading messages from the WebSocket connection
func (fc *FeedConnection) ReadPump(handler *FeedHandler) {
	defer func() {
		handler.UnregisterConnection(fc)
		fc.conn.Close()
	}()

	for {
		var message string
		if err := websocket.Message.Receive(fc.conn, &message); err != nil {
			fc.log.Debug("Read error for user %d: %v", fc.UserID, err)
			break
		}

		// Handle ping/pong
		if message == "ping" {
			if err := fc.writeMessage("pong"); err != nil {
				fc.log.Error("Error sending pong to user %d: %v", fc.UserID, err)
				break
			}
		}
	}
}

func (fc *FeedConnection) writeMessage(message string) error {
	fc.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return websocket.Message.Send(fc.conn, message)
}
