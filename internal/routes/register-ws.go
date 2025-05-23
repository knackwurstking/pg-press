// Routes:
package routes

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func RegisterWS(e *echo.Echo, o *Options) {
	ws := NewWS()

	e.GET(o.ServerPathPrefix+"/htmx/metal-sheets", func(c echo.Context) error {
		websocket.Handler(func(conn *websocket.Conn) {
			defer conn.Close()

			client := NewWSClient("htmx/metal-sheets", conn)
			ws.Register(client)
			defer ws.Unregister(client)

			for {
				// TODO: Send initial data, and remove this data from register-frontend

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

type WS struct {
	Clients []*WSClient

	mutex *sync.Mutex
}

func NewWS() *WS {
	return &WS{
		Clients: make([]*WSClient, 0),
		mutex:   &sync.Mutex{},
	}
}

func (ws *WS) Register(client *WSClient) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !slices.Contains(ws.Clients, client) {
		ws.Clients = append(ws.Clients, client)
	}
}

func (ws *WS) Unregister(client *WSClient) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	newClients := make([]*WSClient, 0)
	for _, c := range ws.Clients {
		if c != client {
			newClients = append(newClients, c)
		}
	}
	ws.Clients = newClients
}

func (ws *WS) Start() {
	// TODO: ...
}

func (ws *WS) Stop() {
	// TODO: ...
}

type WSClient struct {
	Type string
	Conn *websocket.Conn
}

func NewWSClient(t string, client *websocket.Conn) *WSClient {
	return &WSClient{
		Type: t,
		Conn: client,
	}
}
