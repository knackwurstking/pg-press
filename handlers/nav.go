package handlers

import (
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"golang.org/x/net/websocket"

	"github.com/labstack/echo/v4"
)

type Nav struct {
	registry    *services.Registry
	feedHandler *FeedHandler
}

func NewNav(r *services.Registry, fh *FeedHandler) *Nav {
	return &Nav{
		registry:    r,
		feedHandler: fh,
	}
}

func (h *Nav) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/htmx/nav/feed-counter", h.GetFeedCounter),
	})
}

func (h *Nav) GetFeedCounter(c echo.Context) error {
	realIP := c.RealIP()

	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		user, err := utils.GetUserFromContext(c)
		if err != nil {
			slog.Error("WebSocket authentication failed", "real_ip", realIP, "error", err)
			ws.Close()
			return
		}

		slog.Info("WebSocket connection established", "real_ip", realIP, "user_name", user.Name)

		feedConn := h.feedHandler.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			slog.Error("Failed to register WebSocket connection", "real_ip", realIP, "user_name", user.Name)
			ws.Close()
			return
		}

		defer slog.Info("WebSocket connection closed", "real_ip", realIP, "user_name", user.Name)

		go feedConn.WritePump()
		feedConn.ReadPump(h.feedHandler)
	})

	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
