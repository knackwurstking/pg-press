// Package pgvis attachment models.
//
// This file defines the Attachment data structure and its associated
// validation and utility methods. Attachments represent files that can
// be linked to trouble reports or other entities in the system.
//
// Attachments store data in byte form with mime type and unique ID.
package pgvis

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	MinAttachmentIDLength = 1
	MaxAttachmentIDLength = 255
	MaxAttachmentDataSize = 10 * 1024 * 1024 // 10MB
)

// Attachment represents a file attachment with its data and metadata.
type Attachment struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Data     []byte `json:"data"`
}

// Validate checks if the attachment has valid data.
func (a *Attachment) Validate() error {
	if a.ID == "" {
		return NewValidationError("id", "cannot be empty", a.ID)
	}
	if len(a.ID) < MinAttachmentIDLength {
		return NewValidationError("id", "too short", len(a.ID))
	}
	if len(a.ID) > MaxAttachmentIDLength {
		return NewValidationError("id", "too long", len(a.ID))
	}

	if a.MimeType == "" {
		return NewValidationError("mime_type", "cannot be empty", a.MimeType)
	}

	if a.Data == nil {
		return NewValidationError("data", "cannot be nil", a.Data)
	}
	if len(a.Data) > MaxAttachmentDataSize {
		return NewValidationError("data", "too large", len(a.Data))
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

var (
	imageExtensions    = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"}
	documentExtensions = []string{".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt"}
	archiveExtensions  = []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2"}

	mimeTypes = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".rtf":  "application/rtf",
		".odt":  "application/vnd.oasis.opendocument.text",
		".zip":  "application/zip",
		".rar":  "application/vnd.rar",
		".7z":   "application/x-7z-compressed",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".bz2":  "application/x-bzip2",
	}
)

// IsImage checks if the attachment is an image file based on its mime type.
func (a *Attachment) IsImage() bool {
	return strings.HasPrefix(a.MimeType, "image/")
}

// IsDocument checks if the attachment is a document file based on its mime type.
func (a *Attachment) IsDocument() bool {
	documentMimeTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"application/rtf",
		"application/vnd.oasis.opendocument.text",
	}
	return slices.Contains(documentMimeTypes, a.MimeType)
}

// IsArchive checks if the attachment is an archive file based on its mime type.
func (a *Attachment) IsArchive() bool {
	archiveMimeTypes := []string{
		"application/zip",
		"application/vnd.rar",
		"application/x-7z-compressed",
		"application/x-tar",
		"application/gzip",
		"application/x-bzip2",
	}
	return slices.Contains(archiveMimeTypes, a.MimeType)
}

// GetMimeType returns the MIME type of the attachment.
func (a *Attachment) GetMimeType() string {
	return a.MimeType
}

// GetID returns the numeric ID of the attachment.
func (a *Attachment) GetID() int64 {
	// Try to parse the string ID as int64
	if id, err := strconv.ParseInt(a.ID, 10, 64); err == nil {
		return id
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
		return NewValidationError("id", "cannot be empty", newID)
	}
	if len(newID) < MinAttachmentIDLength {
		return NewValidationError("id", "too short", len(newID))
	}
	if len(newID) > MaxAttachmentIDLength {
		return NewValidationError("id", "too long", len(newID))
	}
	a.ID = newID
	return nil
}

// UpdateMimeType updates the attachment's MIME type with validation.
func (a *Attachment) UpdateMimeType(newMimeType string) error {
	newMimeType = strings.TrimSpace(newMimeType)
	if newMimeType == "" {
		return NewValidationError("mime_type", "cannot be empty", newMimeType)
	}
	a.MimeType = newMimeType
	return nil
}

// UpdateData updates the attachment's data with validation.
func (a *Attachment) UpdateData(newData []byte) error {
	if newData == nil {
		return NewValidationError("data", "cannot be nil", newData)
	}
	if len(newData) > MaxAttachmentDataSize {
		return NewValidationError("data", "too large", len(newData))
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
