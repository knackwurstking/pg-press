package htmx

import (
	"net/http"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/webhelpers"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Nav struct {
	DB            *database.DB
	WSFeedHandler *wshandlers.FeedHandler
}

func (h *Nav) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/nav/feed-counter", h.handleFeedCounterWebSocketEcho),
		},
	)
}

// handleFeedCounterWebSocketEcho creates an echo-compatible WebSocket handler
func (h *Nav) handleFeedCounterWebSocketEcho(c echo.Context) error {
	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		// Get user from echo context
		user, err := webhelpers.GetUserFromContext(c)
		if err != nil {
			logger.HTMXHandlerNav().Error("User authentication failed: %#v", err)
			ws.Close()
			return
		}

		logger.HTMXHandlerNav().Info("User authenticated: %s (LastFeed: %d)",
			user.UserName, user.LastFeed)

		// Register the connection with the feed notifier
		feedConn := h.WSFeedHandler.RegisterConnection(
			user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			logger.HTMXHandlerNav().Error(
				"Failed to register connection for user %s", user.UserName)
			ws.Close()
			return
		}

		logger.HTMXHandlerNav().Info("Connection registered for user %s",
			user.UserName)

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(h.WSFeedHandler)

		logger.HTMXHandlerNav().Info("Connection closed for user %s",
			user.UserName)
	})

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
