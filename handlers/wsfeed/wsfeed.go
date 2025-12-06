package wsfeed

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/handlers/wsfeed/templates"
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
)

// Handler manages WebSocket connections for feed counter updates
type Handler struct {
	connections map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	broadcast   chan struct{}
	db          *services.Registry
	mu          sync.RWMutex
}

// NewHandler creates a new feed notification manager
func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		connections: make(map[*Connection]bool),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan struct{}, 100),
		db:          db,
	}
}

// Start begins the notification manager's main loop
func (h *Handler) Start(ctx context.Context) {
	slog.Info("Starting feed notification manager")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down feed notification manager")
			h.closeAllConnections()
			return

		case conn := <-h.register:
			h.registerConnection(conn)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)

		case <-h.broadcast:
			h.broadcastToAllConnections()
		}
	}
}

// RegisterConnection adds a new WebSocket connection to the manager
func (h *Handler) RegisterConnection(userID models.TelegramID, lastFeed models.FeedID, conn *websocket.Conn) *Connection {
	slog.Info("Registering new connection", "telegram_id", userID)
	feedConn := NewConnection(userID, lastFeed, conn)
	h.register <- feedConn
	return feedConn
}

// UnregisterConnection removes a WebSocket connection from the manager
func (h *Handler) UnregisterConnection(conn *Connection) {
	slog.Info("Unregistering connection", "telegram_id", conn.UserID)
	h.unregister <- conn
}

// Broadcast queues a feed counter update for all connected clients
func (h *Handler) Broadcast() {
	select {
	case h.broadcast <- struct{}{}:
		slog.Info("Feed update notification queued")
	default:
		slog.Warn("Broadcast channel full, skipping notification")
	}
}

// ConnectionCount returns the number of active connections
func (h *Handler) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

func (h *Handler) registerConnection(conn *Connection) {
	h.mu.Lock()
	h.connections[conn] = true
	h.mu.Unlock()

	go h.sendUpdate(conn)
}

func (h *Handler) unregisterConnection(conn *Connection) {
	h.mu.Lock()
	if _, ok := h.connections[conn]; ok {
		delete(h.connections, conn)
		close(conn.send)
		close(conn.done)
	}
	h.mu.Unlock()
}

func (h *Handler) broadcastToAllConnections() {
	h.mu.RLock()
	connections := make([]*Connection, 0, len(h.connections))
	for conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	slog.Debug("Broadcasting feed counter update to connections", "connections", len(connections))

	for _, conn := range connections {
		go h.sendUpdate(conn)
	}
}

func (h *Handler) sendUpdate(conn *Connection) {
	html, err := h.renderFeedCounter(conn.LastFeed)
	if err != nil {
		slog.Error("Error rendering feed counter", "telegram_id", conn.UserID, "error", err)
		return
	}

	select {
	case conn.send <- html:
		// Successfully queued
	case <-conn.done:
		// Connection closed
		h.unregisterConnection(conn)
	case <-time.After(writeTimeout):
		slog.Warn("Timeout sending feed counter", "telegram_id", conn.UserID)
	}
}

func (h *Handler) renderFeedCounter(userLastFeed models.FeedID) ([]byte, error) {
	feeds, err := h.db.Feeds.ListRange(0, env.MaxFeedsPerPage)
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
	if err := templates.FeedCounter(count).Render(context.Background(), &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h *Handler) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.connections {
		conn.conn.Close()
		close(conn.send)
		close(conn.done)
		delete(h.connections, conn)
	}

	slog.Info("Closed all WebSocket connections")
}

// Connection represents a WebSocket connection for feed updates
type Connection struct {
	UserID   models.TelegramID
	LastFeed models.FeedID
	conn     *websocket.Conn
	send     chan []byte
	done     chan struct{}
}

// NewConnection creates a new feed connection
func NewConnection(userID models.TelegramID, lastFeed models.FeedID, conn *websocket.Conn) *Connection {
	return &Connection{
		UserID:   userID,
		LastFeed: lastFeed,
		conn:     conn,
		send:     make(chan []byte, 256),
		done:     make(chan struct{}),
	}
}

// WritePump handles writing messages to the WebSocket connection
func (c *Connection) WritePipe() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}

			if err := c.writeMessage(string(message)); err != nil {
				slog.Error("Error writing message", "telegram_id", c.UserID, "error", err)
				return
			}

		case <-ticker.C:
			if err := c.writeMessage("ping"); err != nil {
				slog.Error("Error sending ping", "telegram_id", c.UserID, "error", err)
				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *Connection) ReadPipe(handler *Handler) {
	defer func() {
		handler.UnregisterConnection(c)
		c.conn.Close()
	}()

	for {
		var message string
		if err := websocket.Message.Receive(c.conn, &message); err != nil {
			slog.Debug("Read error", "telegram_id", c.UserID, "error", err)
			break
		}

		// Handle ping/pong
		if message == "ping" {
			if err := c.writeMessage("pong"); err != nil {
				slog.Error("Error sending pong", "telegram_id", c.UserID, "error", err)
				break
			}
		}
	}
}

func (c *Connection) writeMessage(message string) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return websocket.Message.Send(c.conn, message)
}
