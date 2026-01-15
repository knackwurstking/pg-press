package shared

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pg-press/internal/env"
)

var (
	MimeTypes = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
	}
)

type Image struct {
	Name     string `json:"name"`
	Data     []byte `json:"data"`
	MimeType string `json:"mime_type"`
}

func (i *Image) IsImage() bool {
	return strings.HasPrefix(i.MimeType, "image/")
}

func (i *Image) ServerPath() string {
	return fmt.Sprintf("%s/%s", env.ServerPathImages, i.Name)
}
