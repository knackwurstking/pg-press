// API Routes:
//   - GET - "/ws"
package routes

import (
	"log/slog"
	"picow-led/internal/api"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type wsOptions struct {
	ServerPathPrefix string
	WS               *api.WS
}

func wsRoutes(e *echo.Echo, o wsOptions) {
	e.GET(o.ServerPathPrefix+"/ws", func(c echo.Context) error {
		websocket.Handler(func(conn *websocket.Conn) {
			defer conn.Close()

			client := &api.WSClient{
				Conn: conn,
			}
			o.WS.RegisterClient(client)
			defer o.WS.UnregisterClient(client)

			for {
				// Keep alive loop here, this websocket is readonly
				message := ""
				if err := websocket.Message.Receive(conn, &message); err != nil {
					slog.Warn("Receive websocket message", "RealIP", c.RealIP(), "error", err, "path", c.Request().URL.Path)
					break
				}
				slog.Debug("Received websocket message", "RealIP", c.RealIP(), "message", message, "path", c.Request().URL.Path)
			}
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	})
}
