package nav

import (
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/handlers/wsfeed"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
	"golang.org/x/net/websocket"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry    *services.Registry
	feedHandler *wsfeed.Handler
}

func NewHandler(r *services.Registry, fh *wsfeed.Handler) *Handler {
	return &Handler{
		registry:    r,
		feedHandler: fh,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(
			http.MethodGet,
			path+"/feed-counter",
			h.GetFeedCounter,
		),
	})
}

func (h *Handler) GetFeedCounter(c echo.Context) error {
	realIP := c.RealIP()

	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		user, merr := utils.GetUserFromContext(c)
		if merr != nil {
			slog.Error("WebSocket authentication failed", "real_ip", realIP, "error", merr)
			ws.Close()
			return
		}

		feedConn := h.feedHandler.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		if feedConn == nil {
			slog.Error("Failed to register WebSocket connection", "real_ip", realIP, "user_name", user.Name)
			ws.Close()
			return
		}

		defer slog.Info("WebSocket connection closed", "real_ip", realIP, "user_name", user.Name)

		go feedConn.WritePipe()
		feedConn.ReadPipe(h.feedHandler)
	})

	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}
