package api

import (
	"log/slog"
	"sync"
)

type Device struct {
	Server *Server `json:"server" yaml:"server"`

	Online bool   `json:"online" yaml:"-"` //  Not used in configurations
	Error  string `json:"error" yaml:"-"`  //  Not used in configurations

	// Color can be nil
	Color MicroColor `json:"color" yaml:"-"`
	// Pins can be nil
	Pins MicroPins `json:"pins" yaml:"pins"`
}

// Server contains host and port in use from a Device
type Server struct {
	Addr string `json:"addr" yaml:"addr"`
	// Name could be empty (optional)
	Name string `json:"name" yaml:"name"`
}

func GetDevices(devices ...*Device) []*Device {
	wg := &sync.WaitGroup{}
	for _, device := range devices {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r := NewMicroRequest(MicroIDDefault)

			if pins, err := r.Pins(device); err != nil && !device.Online {
				return
			} else {
				device.Pins = pins
			}

			if device.Error != "" {
				slog.Error("Request pins", "error", device.Error, "device.server", device.Server)
			}
			pinsError := device.Error

			if color, err := r.Color(device); err != nil && !device.Online {
				return
			} else {
				device.Color = color
			}

			if device.Error != "" {
				slog.Error("Request color", "error", device.Error, "device.server", device.Server)
			} else {
				device.Error = pinsError
			}
		}()
	}
	wg.Wait()

	return devices
}

func SetColor(c MicroColor, devices ...*Device) []*Device {
	wg := &sync.WaitGroup{}
	for _, d := range devices {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r := NewMicroRequest(MicroIDDefault)
			if err := r.SetColor(d, c); err != nil {
				d.Error = err.Error()
			} else {
				d.Color = c
			}
		}()
	}
	wg.Wait()

	return devices
}

func SetPins(p MicroPins, devices ...*Device) []*Device {
	wg := &sync.WaitGroup{}
	for _, d := range devices {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r := NewMicroRequest(MicroIDDefault)
			if err := r.SetPins(d, p); err != nil {
				d.Error = err.Error()
			} else {
				d.Pins = p
			}
		}()
	}
	wg.Wait()

	return devices
}

//// SetColorForce will set color just like `SetColor`, but will skip the response
//func SetColorForce(o *Config, c MicroColor, devices ...*Device) {
//	wg := &sync.WaitGroup{}
//	for _, d := range devices {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//
//			r := NewMicroRequest(MicroIDNoResponse)
//			r.ID = MicroIDNoResponse
//			if err := r.SetColor(d, c); err != nil {
//				d.Error = err.Error()
//			} else {
//				d.Color = c
//			}
//		}()
//	}
//	wg.Wait()
//}
