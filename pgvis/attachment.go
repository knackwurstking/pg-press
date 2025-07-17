// ai: Organize
package pgvis

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// Attachment represents a file attachment with its metadata
type Attachment struct {
	Name         string `json:"name"`          // Display name of the attachment
	Link         string `json:"link"`          // URL or link to access the attachment
	RelativePath string `json:"relative_path"` // Relative path to the file on the filesystem
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

// Validate checks if the attachment has valid data
func (a *Attachment) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("attachment name cannot be empty")
	}

	if a.Link == "" {
		return fmt.Errorf("attachment link cannot be empty")
	}

	if a.RelativePath == "" {
		return fmt.Errorf("attachment relative path cannot be empty")
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

// UpdateLink updates the attachment's link
func (a *Attachment) UpdateLink(newLink string) {
	a.Link = strings.TrimSpace(newLink)
}

// UpdateName updates the attachment's display name
func (a *Attachment) UpdateName(newName string) {
	a.Name = strings.TrimSpace(newName)
}

// UpdatePath updates the attachment's relative path
func (a *Attachment) UpdatePath(newPath string) {
	a.RelativePath = strings.TrimSpace(newPath)
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
