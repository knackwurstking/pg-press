package api

import (
	"encoding/json"
	"log/slog"
	"slices"
	"sync"

	"golang.org/x/net/websocket"
)

const (
	// BroadcastTypeDevices BroadcastType = "devices"
	BroadcastTypeDevice BroadcastType = "device"
	BroadcastTypeColors BroadcastType = "colors"
)

type BroadcastType string

type BroadcastData struct {
	Type BroadcastType `json:"type"`
	Data any           `json:"data"`
}

type WS struct {
	Clients []*WSClient

	running   bool
	broadcast chan BroadcastData
	done      chan any

	mutex *sync.Mutex
}

func NewWS() *WS {
	return &WS{
		Clients: make([]*WSClient, 0),
		mutex:   &sync.Mutex{},
	}
}

func (ws *WS) RegisterClient(c *WSClient) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if slices.Contains(ws.Clients, c) {
		return
	}

	ws.Clients = append(ws.Clients, c)
}

func (ws *WS) UnregisterClient(c *WSClient) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	for i, client := range ws.Clients {
		if client == c {
			ws.Clients = slices.Delete(ws.Clients, i, i+1)
		}
	}
}

func (ws *WS) Start() error {
	ws.running = true
	defer func() {
		ws.running = false
	}()

	ws.broadcast = make(chan BroadcastData)
	ws.done = make(chan any)

	defer func() {
		close(ws.broadcast)
		close(ws.done)
	}()

	for {
		select {
		case v := <-ws.broadcast:
			{
				wg := &sync.WaitGroup{}
				for _, c := range ws.Clients {
					wg.Add(1)
					go func() {
						defer wg.Done()

						d, err := json.Marshal(v)
						if err != nil {
							slog.Error("Marshal JSON", "error", err, "client", c)
						}

						err = websocket.Message.Send(c.Conn, d)
						if err != nil {
							slog.Warn("Send message", "error", err, "client", c)
						}
					}()
				}
				wg.Wait()
			}
		case <-ws.done:
			break
		}
	}
}

func (ws *WS) Stop() {
	if !ws.running {
		return
	}

	ws.done <- nil
}

func (ws *WS) Broadcast(t BroadcastType, v any) {
	if !ws.running {
		return
	}

	ws.broadcast <- BroadcastData{
		Type: t,
		Data: v,
	}
}

//func (ws *WS) BroadcastDevices(d []*Device) {
//	if !ws.running {
//		return
//	}
//
//	ws.broadcast <- BroadcastData{
//		Type: BroadcastTypeDevices,
//		Data: d,
//	}
//}

func (ws *WS) BroadcastDevice(d *Device) {
	if !ws.running {
		return
	}

	ws.broadcast <- BroadcastData{
		Type: BroadcastTypeDevice,
		Data: d,
	}
}

func (ws *WS) BroadcastColors(c []MicroColor) {
	if !ws.running {
		return
	}

	ws.broadcast <- BroadcastData{
		Type: BroadcastTypeColors,
		Data: c,
	}
}

type WSClient struct {
	Conn *websocket.Conn
}
