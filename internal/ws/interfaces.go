// TODO: Move to interfaces package
package ws

import (
	"context"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"golang.org/x/net/websocket"
)

// WSHandler defines the interface for managing WebSocket connections for feed notifications.
// It provides a contract for handling real-time feed counter updates to connected clients.
// Implementations should manage connection lifecycle, broadcasting updates, and maintaining
// connection state in a thread-safe manner.
//
// Usage:
//   - Call Start() in a goroutine to begin the notification manager
//   - Use RegisterConnection() when a new WebSocket client connects
//   - Call Broadcast() when new feed items are available
//   - The manager handles broadcasting updates to all connected clients
type WSHandler interface {
	// Start begins the notification manager's main loop and should be called in a goroutine.
	// It listens for registration/unregistration events and broadcast requests.
	// The method blocks until the context is cancelled, at which point it gracefully
	// shuts down all connections.
	Start(ctx context.Context)

	// RegisterConnection adds a new WebSocket connection to the manager for the specified user.
	// userID is the unique identifier for the user (typically Telegram ID).
	// lastFeed is the ID of the last feed item the user has seen, used to calculate unread count.
	// conn is the active WebSocket connection.
	// Returns a FeedConnection that manages the individual connection lifecycle.
	RegisterConnection(userID, lastFeed int64, conn *websocket.Conn) *FeedConnection

	// UnregisterConnection removes a WebSocket connection from the manager.
	// This is typically called when a connection is closed or encounters an error.
	// The implementation should clean up associated resources and close channels.
	UnregisterConnection(conn *FeedConnection)

	// Broadcast broadcasts feed counter updates to all connected clients.
	// This should be called whenever new feed items are added to trigger
	// real-time updates to all connected WebSocket clients.
	// The method is non-blocking and queues the notification for processing.
	Broadcast()

	// ConnectionCount returns the current number of active WebSocket connections.
	// This can be used for monitoring and debugging purposes.
	ConnectionCount() int
}

// Connection defines the interface for individual WebSocket connections.
// It abstracts the communication layer for a single WebSocket client connection,
// providing separate read and write pumps for handling bidirectional communication.
// Implementations should handle connection lifecycle, error recovery, and
// graceful shutdown scenarios including browser suspension/resumption.
//
// Usage:
//   - Call WritePump() in a goroutine to handle outgoing messages
//   - Call ReadPump() in the main goroutine to handle incoming messages
//   - ReadPump() typically blocks until the connection is closed
type Connection interface {
	// WritePump handles writing messages to the WebSocket connection.
	// It should be called in a separate goroutine and will run until the connection
	// is closed or an error occurs. The method handles periodic ping messages
	// to keep the connection alive and manages write timeouts for suspended clients.
	WritePump()

	// ReadPump handles reading messages from the WebSocket connection.
	// It processes incoming messages including ping/pong for connection health checks.
	// The method blocks until the connection is closed and automatically handles
	// connection cleanup and unregistration through the provided notifier.
	// notifier is used to unregister the connection when it's closed.
	ReadPump(notifier WSHandler)
}

// Example usage:
//
//	// Create a feed notifier
//	db := database.NewDB(...)
//	templates := embed.FS{...}
//	notifier := NewFeedHandler(db, templates)
//
//	// Start the notifier in a goroutine
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	go notifier.Start(ctx)
//
//	// Register a new WebSocket connection
//	conn := &websocket.Conn{...}
//	feedConn := notifier.RegisterConnection(userID, lastFeedID, conn)
//
//	// Start the connection pumps
//	go feedConn.WritePump()
//	feedConn.ReadPump(notifier) // blocks until connection closes
//
//	// Notify about new feeds (from another goroutine)
//	notifier.Broadcast()

// Compile-time interface compliance checks
var (
	_ WSHandler              = (*FeedHandler)(nil)
	_ interfaces.Broadcaster = (*FeedHandler)(nil)
	_ Connection             = (*FeedConnection)(nil)
)
