// Routes:
package routes

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func RegisterWS(e *echo.Echo, o *Options) {
	e.GET(o.ServerPathPrefix+"/htmx/metal-sheets", func(c echo.Context) error {
		websocket.Handler(func(conn *websocket.Conn) {
			// TODO: Register client, need a websocket handler first

			for {
				// TODO: Read, parse and handle client data
				message := ""
				if err := websocket.Message.Receive(conn, &message); err != nil {
					slog.Warn("Receive websocket message",
						"error", err,
						"RealIP", c.RealIP(),
						"path", c.Request().URL.Path)
					break
				}
				slog.Debug("Received websocket message",
					"message", message,
					"RealIP", c.RealIP(),
					"path", c.Request().URL.Path)
			}
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	})
}
