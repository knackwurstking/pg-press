package api

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Devices []*Device `json:"devices,omitempty" yaml:"devices,omitempty"`
}

func GetConfig(paths ...string) (*Config, error) {
	o := &Config{
		Devices: []*Device{},
	}

	defer func() {
		wg := &sync.WaitGroup{}
		for _, d := range o.Devices {
			wg.Add(1)
			go func() {
				defer wg.Done()

				if d.Pins != nil {
					if len(d.Pins) > 0 {
						SetPins(d.Pins, d)
					}
				}
			}()
		}
		wg.Wait()
	}()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}
		f, err := os.Open(absPath)
		if err != nil {
			continue
		}
		d, err := io.ReadAll(f)
		if err != nil {
			return o, err
		}
		err = yaml.Unmarshal(d, o)
		if err != nil {
			return o, err
		}
	}

	return o, nil
}
