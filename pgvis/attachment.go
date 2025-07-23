// Package pgvis attachment models.
//
// This file defines the Attachment data structure and its associated
// validation and utility methods. Attachments represent files that can
// be linked to trouble reports or other entities in the system.
//
// TODO:
//   - Change attachments to store data in byte form and the mime type
//   - Unique: ID, Path
//   - Other Fields: MimeType string, Data []byte
package pgvis

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

const (
	MinAttachmentNameLength = 1
	MaxAttachmentNameLength = 255
	MaxAttachmentPathLength = 1000
)

// Attachment represents a file attachment with its metadata.
type Attachment struct {
	Name         string `json:"name"`
	Link         string `json:"link"`
	RelativePath string `json:"relative_path"`
}

// Validate checks if the attachment has valid data.
func (a *Attachment) Validate() error {
	if a.Name == "" {
		return NewValidationError("name", "cannot be empty", a.Name)
	}
	if len(a.Name) < MinAttachmentNameLength {
		return NewValidationError("name", "too short", len(a.Name))
	}
	if len(a.Name) > MaxAttachmentNameLength {
		return NewValidationError("name", "too long", len(a.Name))
	}

	if a.Link == "" {
		return NewValidationError("link", "cannot be empty", a.Link)
	}

	if a.RelativePath == "" {
		return NewValidationError("relative_path", "cannot be empty", a.RelativePath)
	}
	if len(a.RelativePath) > MaxAttachmentPathLength {
		return NewValidationError("relative_path", "too long", len(a.RelativePath))
	}

	return nil
}

// GetFileExtension returns the file extension of the attachment.
func (a *Attachment) GetFileExtension() string {
	return filepath.Ext(a.Name)
}

// GetFileName returns the filename without extension.
func (a *Attachment) GetFileName() string {
	ext := filepath.Ext(a.Name)
	return strings.TrimSuffix(a.Name, ext)
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

// IsImage checks if the attachment is an image file based on its extension.
func (a *Attachment) IsImage() bool {
	ext := strings.ToLower(a.GetFileExtension())
	return slices.Contains(imageExtensions, ext)
}

// IsDocument checks if the attachment is a document file based on its extension.
func (a *Attachment) IsDocument() bool {
	ext := strings.ToLower(a.GetFileExtension())
	return slices.Contains(documentExtensions, ext)
}

// IsArchive checks if the attachment is an archive file based on its extension.
func (a *Attachment) IsArchive() bool {
	ext := strings.ToLower(a.GetFileExtension())
	return slices.Contains(archiveExtensions, ext)
}

// GetMimeType returns the MIME type based on the file extension.
func (a *Attachment) GetMimeType() string {
	ext := strings.ToLower(a.GetFileExtension())
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

// String returns a string representation of the attachment.
func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{Name: %s, Link: %s, Path: %s}",
		a.Name, a.Link, a.RelativePath)
}

// Clone creates a deep copy of the attachment.
func (a *Attachment) Clone() *Attachment {
	return &Attachment{
		Name:         a.Name,
		Link:         a.Link,
		RelativePath: a.RelativePath,
	}
}

// UpdateLink updates the attachment's link with validation.
func (a *Attachment) UpdateLink(newLink string) error {
	newLink = strings.TrimSpace(newLink)
	if newLink == "" {
		return NewValidationError("link", "cannot be empty", newLink)
	}
	a.Link = newLink
	return nil
}

// UpdateName updates the attachment's display name with validation.
func (a *Attachment) UpdateName(newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return NewValidationError("name", "cannot be empty", newName)
	}
	if len(newName) < MinAttachmentNameLength {
		return NewValidationError("name", "too short", len(newName))
	}
	if len(newName) > MaxAttachmentNameLength {
		return NewValidationError("name", "too long", len(newName))
	}
	a.Name = newName
	return nil
}

// UpdatePath updates the attachment's relative path with validation.
func (a *Attachment) UpdatePath(newPath string) error {
	newPath = strings.TrimSpace(newPath)
	if newPath == "" {
		return NewValidationError("relative_path", "cannot be empty", newPath)
	}
	if len(newPath) > MaxAttachmentPathLength {
		return NewValidationError("relative_path", "too long", len(newPath))
	}
	a.RelativePath = newPath
	return nil
}

// Equals checks if two attachments are equal.
func (a *Attachment) Equals(other *Attachment) bool {
	if other == nil {
		return false
	}
	return a.Name == other.Name &&
		a.Link == other.Link &&
		a.RelativePath == other.RelativePath
}
