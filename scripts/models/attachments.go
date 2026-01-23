package models

var (
	mimeTypes = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
	}
)

type Attachment struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Data     []byte `json:"data"`
}

func (a *Attachment) GetExtension() string {
	for ext, mimeType := range mimeTypes {
		if mimeType == a.MimeType {
			return ext
		}
	}
	return ""
}
