package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"

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
}

// NewFeedHandler creates a new feed notification manager
func NewFeedHandler(db *services.Registry) *FeedHandler {
	return &FeedHandler{
		connections: make(map[*FeedConnection]bool),
		register:    make(chan *FeedConnection),
		unregister:  make(chan *FeedConnection),
		broadcast:   make(chan struct{}, 100),
		db:          db,
	}
}

// Start begins the notification manager's main loop
func (fh *FeedHandler) Start(ctx context.Context) {
	slog.Info("Starting feed notification manager")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down feed notification manager")
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
func (fh *FeedHandler) RegisterConnection(userID models.TelegramID, lastFeed models.FeedID, conn *websocket.Conn) *FeedConnection {
	slog.Info("Registering new connection", "telegram_id", userID)
	feedConn := NewFeedConnection(userID, lastFeed, conn)
	fh.register <- feedConn
	return feedConn
}

// UnregisterConnection removes a WebSocket connection from the manager
func (fh *FeedHandler) UnregisterConnection(conn *FeedConnection) {
	slog.Info("Unregistering connection", "telegram_id", conn.UserID)
	fh.unregister <- conn
}

// Broadcast queues a feed counter update for all connected clients
func (fh *FeedHandler) Broadcast() {
	select {
	case fh.broadcast <- struct{}{}:
		slog.Debug("Feed update notification queued")
	default:
		slog.Warn("Broadcast channel full, skipping notification")
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
}

func (fh *FeedHandler) broadcastToAllConnections() {
	fh.mu.RLock()
	connections := make([]*FeedConnection, 0, len(fh.connections))
	for conn := range fh.connections {
		connections = append(connections, conn)
	}
	fh.mu.RUnlock()

	slog.Debug("Broadcasting feed counter update to connections", "connections", len(connections))

	for _, conn := range connections {
		go fh.sendUpdate(conn)
	}
}

func (fh *FeedHandler) sendUpdate(conn *FeedConnection) {
	html, err := fh.renderFeedCounter(conn.LastFeed)
	if err != nil {
		slog.Error("Error rendering feed counter", "telegram_id", conn.UserID, "error", err)
		return
	}

	select {
	case conn.send <- html:
		// Successfully queued
	case <-conn.done:
		// Connection closed
		fh.UnregisterConnection(conn)
	case <-time.After(writeTimeout):
		slog.Warn("Timeout sending feed counter", "telegram_id", conn.UserID)
	}
}

func (fh *FeedHandler) renderFeedCounter(userLastFeed models.FeedID) ([]byte, error) {
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

	slog.Info("Closed all WebSocket connections")
}

// FeedConnection represents a WebSocket connection for feed updates
type FeedConnection struct {
	UserID   models.TelegramID
	LastFeed models.FeedID
	conn     *websocket.Conn
	send     chan []byte
	done     chan struct{}
}

// NewFeedConnection creates a new feed connection
func NewFeedConnection(userID models.TelegramID, lastFeed models.FeedID, conn *websocket.Conn) *FeedConnection {
	return &FeedConnection{
		UserID:   userID,
		LastFeed: lastFeed,
		conn:     conn,
		send:     make(chan []byte, 256),
		done:     make(chan struct{}),
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
				slog.Error("Error writing message", "telegram_id", fc.UserID, "error", err)
				return
			}

		case <-ticker.C:
			if err := fc.writeMessage("ping"); err != nil {
				slog.Error("Error sending ping", "telegram_id", fc.UserID, "error", err)
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
			slog.Debug("Read error", "telegram_id", fc.UserID, "error", err)
			break
		}

		// Handle ping/pong
		if message == "ping" {
			if err := fc.writeMessage("pong"); err != nil {
				slog.Error("Error sending pong", "telegram_id", fc.UserID, "error", err)
				break
			}
		}
	}
}

func (fc *FeedConnection) writeMessage(message string) error {
	fc.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return websocket.Message.Send(fc.conn, message)
}
