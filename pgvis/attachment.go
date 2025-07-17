// Package pgvis attachment models.
//
// This file defines the Attachment data structure and its associated
// validation and utility methods. Attachments represent files that can
// be linked to trouble reports or other entities in the system.
package pgvis

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

const (
	// Validation constants for attachments
	MinAttachmentNameLength = 1
	MaxAttachmentNameLength = 255
	MaxAttachmentPathLength = 1000
)

// Attachment represents a file attachment with its metadata.
// It contains information about files that can be linked to trouble reports
// or other entities in the system.
type Attachment struct {
	// Name is the display name of the attachment
	Name string `json:"name"`
	// Link is the URL or link to access the attachment
	Link string `json:"link"`
	// RelativePath is the relative path to the file on the filesystem
	RelativePath string `json:"relative_path"`
}

// NewAttachment creates a new attachment with the provided details
func NewAttachment(name, link, relativePath string) *Attachment {
	return &Attachment{
		Name:         strings.TrimSpace(name),
		Link:         strings.TrimSpace(link),
		RelativePath: strings.TrimSpace(relativePath),
	}
}

// NewAttachmentFromPath creates a new attachment from a file path
// The name is derived from the filename, and the link is constructed from the relative path
func NewAttachmentFromPath(relativePath string) *Attachment {
	name := filepath.Base(relativePath)
	link := fmt.Sprintf("/attachments/%s", relativePath)

	return &Attachment{
		Name:         name,
		Link:         link,
		RelativePath: relativePath,
	}
}

// Validate checks if the attachment has valid data.
//
// Returns:
//   - error: MultiError containing all validation failures, or nil if valid
func (a *Attachment) Validate() error {
	multiErr := NewMultiError()

	// Validate name
	if a.Name == "" {
		multiErr.Add(NewValidationError("name", "cannot be empty", a.Name))
	} else {
		if len(a.Name) < MinAttachmentNameLength {
			multiErr.Add(NewValidationError("name", "too short", len(a.Name)))
		}
		if len(a.Name) > MaxAttachmentNameLength {
			multiErr.Add(NewValidationError("name", "too long", len(a.Name)))
		}
	}

	// Validate link
	if a.Link == "" {
		multiErr.Add(NewValidationError("link", "cannot be empty", a.Link))
	}

	// Validate relative path
	if a.RelativePath == "" {
		multiErr.Add(NewValidationError("relative_path", "cannot be empty", a.RelativePath))
	} else {
		if len(a.RelativePath) > MaxAttachmentPathLength {
			multiErr.Add(NewValidationError("relative_path", "too long", len(a.RelativePath)))
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}

	return nil
}

// GetFileExtension returns the file extension of the attachment
func (a *Attachment) GetFileExtension() string {
	return filepath.Ext(a.Name)
}

// GetFileName returns the filename without extension
func (a *Attachment) GetFileName() string {
	ext := filepath.Ext(a.Name)
	return strings.TrimSuffix(a.Name, ext)
}

// IsImage checks if the attachment is an image file based on its extension
func (a *Attachment) IsImage() bool {
	ext := strings.ToLower(a.GetFileExtension())
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"}

	return slices.Contains(imageExtensions, ext)
}

// IsDocument checks if the attachment is a document file based on its extension
func (a *Attachment) IsDocument() bool {
	ext := strings.ToLower(a.GetFileExtension())
	documentExtensions := []string{".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt"}

	return slices.Contains(documentExtensions, ext)
}

// IsArchive checks if the attachment is an archive file based on its extension
func (a *Attachment) IsArchive() bool {
	ext := strings.ToLower(a.GetFileExtension())
	archiveExtensions := []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2"}

	return slices.Contains(archiveExtensions, ext)
}

// GetMimeType returns the MIME type based on the file extension
func (a *Attachment) GetMimeType() string {
	ext := strings.ToLower(a.GetFileExtension())

	mimeTypes := map[string]string{
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

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// String returns a string representation of the attachment
func (a *Attachment) String() string {
	return fmt.Sprintf("Attachment{Name: %s, Link: %s, Path: %s}",
		a.Name, a.Link, a.RelativePath)
}

// Clone creates a deep copy of the attachment
func (a *Attachment) Clone() *Attachment {
	return &Attachment{
		Name:         a.Name,
		Link:         a.Link,
		RelativePath: a.RelativePath,
	}
}

// UpdateLink updates the attachment's link with validation.
//
// Parameters:
//   - newLink: The new link for the attachment
//
// Returns:
//   - error: Validation error if the link is invalid
func (a *Attachment) UpdateLink(newLink string) error {
	newLink = strings.TrimSpace(newLink)

	if newLink == "" {
		return NewValidationError("link", "cannot be empty", newLink)
	}

	a.Link = newLink
	return nil
}

// UpdateName updates the attachment's display name with validation.
//
// Parameters:
//   - newName: The new display name for the attachment
//
// Returns:
//   - error: Validation error if the name is invalid
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
//
// Parameters:
//   - newPath: The new relative path for the attachment
//
// Returns:
//   - error: Validation error if the path is invalid
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

// Equals checks if two attachments are equal
func (a *Attachment) Equals(other *Attachment) bool {
	if other == nil {
		return false
	}

	return a.Name == other.Name &&
		a.Link == other.Link &&
		a.RelativePath == other.RelativePath
}
