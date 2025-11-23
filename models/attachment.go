package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/errors"
)

const (
	// TODO: Move to "env" package
	MinIDLength = 1
	MaxIDLength = 255
	MaxDataSize = 10 * 1024 * 1024 // 10MB
)

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

type AttachmentID int64

// Attachment represents a file attachment with its data and metadata.
type Attachment struct {
	ID       string `json:"id"` // TODO: Migrate this to a numeric ID, also need to update the database table for this
	MimeType string `json:"mime_type"`
	Data     []byte `json:"data"`
}

// Validate checks if the attachment has valid data.
func (a *Attachment) Validate() error {
	if a.MimeType == "" {
		return errors.NewValidationError("mime_type cannot be empty")
	}

	if !a.IsImage() {
		return errors.NewValidationError("mime_type: only image files are allowed")
	}

	if a.Data == nil {
		return errors.NewValidationError("data cannot be nil")
	}

	if len(a.Data) > MaxDataSize {
		return errors.NewValidationError("data too large")
	}

	return nil
}

// GetFileExtension returns the file extension based on the mime type.
func (a *Attachment) GetFileExtension() string {
	for ext, mimeType := range mimeTypes {
		if mimeType == a.MimeType {
			return ext
		}
	}
	return ""
}

// IsImage checks if the attachment is an image file based on its mime type.
func (a *Attachment) IsImage() bool {
	return strings.HasPrefix(a.MimeType, "image/")
}

// GetMimeType returns the MIME type of the attachment.
func (a *Attachment) GetMimeType() string {
	return a.MimeType
}

// GetID returns the numeric ID of the attachment.
func (a *Attachment) GetID() AttachmentID {
	// Try to parse the string ID as int64
	if id, err := strconv.ParseInt(a.ID, 10, 64); err == nil {
		return AttachmentID(id)
	}
	return 0 // Return 0 for invalid IDs
}

// String returns a string representation of the attachment.
func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{ID: %s, MimeType: %s, DataSize: %d}",
		a.ID, a.MimeType, len(a.Data))
}
