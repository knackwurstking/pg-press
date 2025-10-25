// TODO: Remove useless stuff
package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/errors"
)

const (
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

// Clone creates a deep copy of the attachment.
func (a *Attachment) Clone() *Attachment {
	dataCopy := make([]byte, len(a.Data))
	copy(dataCopy, a.Data)
	return &Attachment{
		ID:       a.ID,
		MimeType: a.MimeType,
		Data:     dataCopy,
	}
}

// UpdateID updates the attachment's ID with validation.
func (a *Attachment) UpdateID(newID string) error {
	newID = strings.TrimSpace(newID)

	if newID == "" {
		return errors.NewValidationError("id cannot be empty")
	}

	if len(newID) < MinIDLength {
		return errors.NewValidationError("id too short")
	}

	if len(newID) > MaxIDLength {
		return errors.NewValidationError("id too long")
	}

	a.ID = newID
	return nil
}

// UpdateMimeType updates the attachment's MIME type with validation.
func (a *Attachment) UpdateMimeType(newMimeType string) error {
	newMimeType = strings.TrimSpace(newMimeType)
	if newMimeType == "" {
		return errors.NewValidationError("mime_type cannot be empty")
	}
	a.MimeType = newMimeType
	return nil
}

// UpdateData updates the attachment's data with validation.
func (a *Attachment) UpdateData(newData []byte) error {
	if newData == nil {
		return errors.NewValidationError("data cannot be nil")
	}

	if len(newData) > MaxDataSize {
		return errors.NewValidationError("data too large")
	}

	a.Data = make([]byte, len(newData))
	copy(a.Data, newData)
	return nil
}

// Equals checks if two attachments are equal.
func (a *Attachment) Equals(other *Attachment) bool {
	if other == nil {
		return false
	}

	if a.ID != other.ID || a.MimeType != other.MimeType {
		return false
	}

	if len(a.Data) != len(other.Data) {
		return false
	}

	for i := range a.Data {
		if a.Data[i] != other.Data[i] {
			return false
		}
	}

	return true
}
