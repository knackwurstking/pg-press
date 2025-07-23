package nav

import (
	"io/fs"
	"log"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/internal/notifications"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
	feedNotifier     *notifications.FeedNotifier
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS, feedNotifier *notifications.FeedNotifier) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
		feedNotifier:     feedNotifier,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// This approach keeps the WebSocket handler within Echo's middleware chain
	// which is better for authentication that depends on Echo's context
	e.GET(h.serverPathPrefix+"/nav/feed-counter", h.handleFeedCounterWebSocketEcho)
}

// handleFeedCounterWebSocketEcho creates an echo-compatible WebSocket handler
func (h *Handler) handleFeedCounterWebSocketEcho(c echo.Context) error {
	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		// Get user from echo context
		user, herr := utils.GetUserFromContext(c)
		if herr != nil {
			log.Printf("[WebSocket] User authentication failed: %v", herr)
			ws.Close()
			return
		}

		log.Printf("[WebSocket] User authenticated: %s (LastFeed: %d)", user.UserName, user.LastFeed)

		// Register the connection with the feed notifier
		feedConn := h.feedNotifier.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			log.Printf("[WebSocket] Failed to register connection for user %s", user.UserName)
			ws.Close()
			return
		}

		log.Printf("[WebSocket] Connection registered for user %s", user.UserName)

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(h.feedNotifier)

		log.Printf("[WebSocket] Connection closed for user %s", user.UserName)
	})

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
