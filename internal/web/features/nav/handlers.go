package nav

import (
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers"
	"github.com/knackwurstking/pgpress/pkg/logger"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Handler struct {
	*handlers.BaseHandler

	wsFeedHandler *wshandlers.FeedHandler
}

func NewHandler(db *services.Registry, ws *wshandlers.FeedHandler) *Handler {
	return &Handler{
		BaseHandler:   handlers.NewBaseHandler(db, logger.NewComponentLogger("Nav")),
		wsFeedHandler: ws,
	}
}

func (h *Handler) GetFeedCounter(c echo.Context) error {
	remoteIP := c.RealIP()

	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		// Get user from echo context
		user, err := h.GetUserFromContext(c)
		if err != nil {
			h.Log.Error("WebSocket authentication failed from %s: %v", remoteIP, err)
			ws.Close()
			return
		}

		h.Log.Info("WebSocket connection established for user %s from %s", user.Name, remoteIP)

		// Register the connection with the feed notifier
		feedConn := h.wsFeedHandler.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			h.Log.Error("Failed to register WebSocket connection for user %s from %s",
				user.Name, remoteIP)
			ws.Close()
			return
		}

		// Track active connections
		defer func() {
			h.Log.Info("WebSocket connection closed for user %s from %s",
				user.Name, remoteIP)
		}()

		// Start the write pump in a goroutine
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		feedConn.ReadPump(h.wsFeedHandler)
	})

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())

	return nil
}
