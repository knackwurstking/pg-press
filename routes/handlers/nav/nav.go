package nav

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/internal/notifications"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
	upgrader         websocket.Upgrader
	feedNotifier     *notifications.FeedNotifier
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS, feedNotifier *notifications.FeedNotifier) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
		feedNotifier:     feedNotifier,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/nav/feed-counter", h.handleFeedCounterWebSocket)
}

func (h *Handler) handleFeedCounterWebSocket(c echo.Context) error {
	// Upgrade the HTTP connection to WebSocket
	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade failed for %s: %v", c.RealIP(), err)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("websocket upgrade failed: %w", err),
		)
	}
	defer ws.Close()

	// Get user from context before starting WebSocket loop
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		log.Printf("[WebSocket] User authentication failed for %s: %v", c.RealIP(), herr)
		return herr
	}

	log.Printf("[WebSocket] User authenticated: %s (LastFeed: %d)", user.UserName, user.LastFeed)

	// Register the connection with the feed notifier
	feedConn := h.feedNotifier.RegisterConnection(user.TelegramID, user.LastFeed, ws)
	if feedConn == nil {
		log.Printf("[WebSocket] Failed to register connection for user %s", user.UserName)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"failed to register websocket connection",
		)
	}

	log.Printf("[WebSocket] Connection registered for user %s", user.UserName)

	// Start the write pump in a goroutine
	go feedConn.WritePump()

	// Start the read pump (this will block until connection closes)
	feedConn.ReadPump(h.feedNotifier)

	log.Printf("[WebSocket] Connection closed for user %s", user.UserName)
	return nil
}
