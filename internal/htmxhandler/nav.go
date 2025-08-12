package htmxhandler

import (
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/knackwurstking/pgpress/internal/wshandler"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Nav struct {
	*Base
	WSHandler *wshandler.WSHandlers
}

func (h *Nav) RegisterRoutes(e *echo.Echo) {
	// This approach keeps the WebSocket handler within Echo's middleware chain
	// which is better for authentication that depends on Echo's context
	e.GET(h.ServerPathPrefix+"/feed-counter",
		h.handleFeedCounterWebSocketEcho)
}

// handleFeedCounterWebSocketEcho creates an echo-compatible WebSocket handler
func (h *Nav) handleFeedCounterWebSocketEcho(c echo.Context) error {
	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		// Get user from echo context
		user, err := utils.GetUserFromContext(c)
		if err != nil {
			logger.WebSocket().Error("User authentication failed: %#v", err)
			ws.Close()
			return
		}

		logger.WebSocket().Info("User authenticated: %s (LastFeed: %d)",
			user.UserName, user.LastFeed)

		// Register the connection with the feed notifier
		feedConn := h.WSHandler.Feed.RegisterConnection(
			user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			logger.WebSocket().Error(
				"Failed to register connection for user %s", user.UserName)
			ws.Close()
			return
		}

		logger.WebSocket().Info("Connection registered for user %s",
			user.UserName)

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(h.WSHandler.Feed)

		logger.WebSocket().Info("Connection closed for user %s",
			user.UserName)
	})

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
