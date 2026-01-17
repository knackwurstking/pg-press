package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knackwurstking/pg-press/internal/env"
)

type Image struct {
	Name string
	Data []byte
}

func NewImage(name string, data []byte) *Image {
	return &Image{
		Name: name,
		Data: data,
	}
}

func (i *Image) Path() string {
	return filepath.Join(env.ServerPathImages, i.Name)
}

func (i *Image) ReadFile() error {
	var err error
	i.Data, err = os.ReadFile(i.Path())
	if err != nil {
		return fmt.Errorf("read image file %s: %w", i.Path(), err)
	}
	return nil
}

func (i *Image) MimeType() string {
	switch filepath.Ext(i.Name) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
