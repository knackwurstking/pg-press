package nav

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Nav struct {
	*handlers.BaseHandler

	WSFeedHandler *wshandlers.FeedHandler
}

func NewNav(db *database.DB, wsFeedHandler *wshandlers.FeedHandler) *Nav {
	return &Nav{
		BaseHandler:   handlers.NewBaseHandler(db, logger.HTMXHandlerNav()),
		WSFeedHandler: wsFeedHandler,
	}
}

func (h *Nav) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/nav/feed-counter", h.HandleFeedCounterWebSocketEcho),
		},
	)
}

// handleFeedCounterWebSocketEcho creates an echo-compatible WebSocket handler
func (h *Nav) HandleFeedCounterWebSocketEcho(c echo.Context) error {
	remoteIP := c.RealIP()

	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		// Get user from echo context
		user, err := h.GetUserFromContext(c)
		if err != nil {
			h.LogError("WebSocket authentication failed from %s: %v", remoteIP, err)
			ws.Close()
			return
		}

		h.LogInfo("WebSocket connection established for user %s from %s", user.Name, remoteIP)

		// Register the connection with the feed notifier
		feedConn := h.WSFeedHandler.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			h.LogError("Failed to register WebSocket connection for user %s from %s",
				user.Name, remoteIP)
			ws.Close()
			return
		}

		// Track active connections
		defer func() {
			h.LogInfo("WebSocket connection closed for user %s from %s",
				user.Name, remoteIP)
		}()

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(h.WSFeedHandler)
	})

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())

	return nil
}
