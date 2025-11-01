package handlers

import (
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
	remoteIP := c.RealIP()

	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		user, err := GetUserFromContext(c)
		if err != nil {
			h.Log.Error("WebSocket authentication failed from %s: %v", remoteIP, err)
			ws.Close()
			return
		}

		h.Log.Info("WebSocket connection established for user %s from %s", user.Name, remoteIP)

		feedConn := h.feedHandler.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			h.Log.Error("Failed to register WebSocket connection for user %s from %s", user.Name, remoteIP)
			ws.Close()
			return
		}

		defer h.Log.Info("WebSocket connection closed for user %s from %s", user.Name, remoteIP)

		go feedConn.WritePump()
		feedConn.ReadPump(h.feedHandler)
	})

	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
