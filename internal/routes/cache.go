package routes

import (
	"fmt"
	"picow-led/internal/api"
	"slices"
	"sync"
)

var cache *Cache

type Cache struct {
	devices []*api.Device
	colors  []api.MicroColor

	mutex *sync.Mutex
	ws    *api.WS
}

func NewCache(ws *api.WS) *Cache {
	return &Cache{
		devices: make([]*api.Device, 0),
		colors: []api.MicroColor{
			{255, 255, 255, 255},
			{255, 0, 0, 0},
			{0, 255, 0, 0},
			{0, 0, 255, 0},
		},

		mutex: &sync.Mutex{},
		ws:    ws,
	}
}

func (c *Cache) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.ws != nil {
		c.ws.Stop()
	}
}

func (c *Cache) Devices() []*api.Device {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.devices
}

func (c *Cache) SetDevices(devices ...*api.Device) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.devices = devices
}

func (c *Cache) UpdateDevice(addr string, device *api.Device) (*api.Device, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, d := range c.devices {
		if d.Server.Addr != addr {
			continue
		}

		// Only merge things changed after PostDevicesColor call
		d.Color = device.Color
		d.Error = device.Error
		d.Online = device.Online

		if c.ws != nil {
			c.ws.BroadcastDevice(d)
		}

		return d, nil
	}

	return nil, fmt.Errorf("device \"%s\" not found", addr)
}

func (c *Cache) Colors() []api.MicroColor {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.colors
}

func (c *Cache) UpdateColor(index int, color api.MicroColor) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.colors)-1 < index {
		return fmt.Errorf("no such index: %d", index)
	}

	if index < 0 {
		c.colors = append(c.colors, color)
	} else {
		c.colors[index] = color
	}

	if c.ws != nil {
		c.ws.BroadcastColors(c.colors)
	}

	return nil
}

func (c *Cache) DeleteColor(index int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.colors = slices.Delete(c.colors, index, index+1)

	if c.ws != nil {
		go c.ws.BroadcastColors(c.colors)
	}

	return nil
}
