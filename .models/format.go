package models

import "fmt"

type Format struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (f Format) String() string {
	if f.Width == 0 && f.Height == 0 {
		return ""
	}

	return fmt.Sprintf("%dx%d", f.Width, f.Height)
}
