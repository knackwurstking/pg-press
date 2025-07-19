// Package pgvis provides trouble report models for tracking issues and problems.
package pgvis

import (
	"fmt"
	"strings"
)

const (
	MinTitleLength   = 1
	MaxTitleLength   = 500
	MinContentLength = 1
	MaxContentLength = 50000
)

// TroubleReport represents a trouble report in the system.
type TroubleReport struct {
	ID                int                       `json:"id"`
	Title             string                    `json:"title"`
	Content           string                    `json:"content"`
	LinkedAttachments []*Attachment             `json:"linked_attachments"`
	Modified          *Modified[*TroubleReport] `json:"modified"`
}

// NewTroubleReport creates a new trouble report with the provided details.
func NewTroubleReport(m *Modified[*TroubleReport], title, content string) *TroubleReport {
	return &TroubleReport{
		Title:             strings.TrimSpace(title),
		Content:           strings.TrimSpace(content),
		LinkedAttachments: make([]*Attachment, 0),
		Modified:          m,
	}
}

// NewBasicTroubleReport creates a trouble report with minimal metadata.
func NewBasicTroubleReport(title, content string) *TroubleReport {
	systemUser := &User{UserName: "system"}
	modified := NewModified(systemUser, (*TroubleReport)(nil))
	return NewTroubleReport(modified, title, content)
}

// Validate checks if the trouble report has valid data.
func (tr *TroubleReport) Validate() error {
	if err := tr.validateTitle(tr.Title); err != nil {
		return err
	}
	if err := tr.validateContent(tr.Content); err != nil {
		return err
	}
	return tr.validateAttachments()
}

func (tr *TroubleReport) validateTitle(title string) error {
	if title == "" {
		return NewValidationError("title", "cannot be empty", title)
	}
	if len(title) < MinTitleLength {
		return NewValidationError("title", "too short", len(title))
	}
	if len(title) > MaxTitleLength {
		return NewValidationError("title", "too long", len(title))
	}
	return nil
}

func (tr *TroubleReport) validateContent(content string) error {
	if content == "" {
		return NewValidationError("content", "cannot be empty", content)
	}
	if len(content) < MinContentLength {
		return NewValidationError("content", "too short", len(content))
	}
	if len(content) > MaxContentLength {
		return NewValidationError("content", "too long", len(content))
	}
	return nil
}

func (tr *TroubleReport) validateAttachments() error {
	for i, attachment := range tr.LinkedAttachments {
		if attachment == nil {
			return NewValidationError("linked_attachments", "attachment cannot be nil", i)
		}
	}
	return nil
}

// UpdateTitle updates the trouble report title with validation.
func (tr *TroubleReport) UpdateTitle(newTitle string) error {
	newTitle = strings.TrimSpace(newTitle)
	if err := tr.validateTitle(newTitle); err != nil {
		return err
	}
	tr.Title = newTitle
	return nil
}

// UpdateContent updates the trouble report content with validation.
func (tr *TroubleReport) UpdateContent(newContent string) error {
	newContent = strings.TrimSpace(newContent)
	if err := tr.validateContent(newContent); err != nil {
		return err
	}
	tr.Content = newContent
	return nil
}

// AddAttachment adds a new attachment to the trouble report.
func (tr *TroubleReport) AddAttachment(attachment *Attachment) error {
	if attachment == nil {
		return NewValidationError("attachment", "cannot be nil", nil)
	}
	if tr.LinkedAttachments == nil {
		tr.LinkedAttachments = make([]*Attachment, 0)
	}
	tr.LinkedAttachments = append(tr.LinkedAttachments, attachment)
	return nil
}

// RemoveAttachment removes an attachment by index.
func (tr *TroubleReport) RemoveAttachment(index int) error {
	if len(tr.LinkedAttachments) == 0 {
		return NewValidationError("index", "no attachments to remove", index)
	}
	if index < 0 || index >= len(tr.LinkedAttachments) {
		return NewValidationError("index", "out of range", index)
	}
	tr.LinkedAttachments = append(tr.LinkedAttachments[:index], tr.LinkedAttachments[index+1:]...)
	return nil
}

// HasAttachments returns true if the trouble report has any attachments.
func (tr *TroubleReport) HasAttachments() bool {
	return len(tr.LinkedAttachments) > 0
}

// AttachmentCount returns the number of attachments.
func (tr *TroubleReport) AttachmentCount() int {
	return len(tr.LinkedAttachments)
}

// GetSummary returns a brief summary of the trouble report for display purposes.
func (tr *TroubleReport) GetSummary(maxLength int) string {
	if maxLength <= 0 {
		return ""
	}
	content := strings.TrimSpace(tr.Content)
	if len(content) <= maxLength {
		return content
	}
	if maxLength <= 3 {
		return content[:maxLength]
	}
	return content[:maxLength-3] + "..."
}

// Clone creates a deep copy of the trouble report.
func (tr *TroubleReport) Clone() *TroubleReport {
	clone := &TroubleReport{
		ID:      tr.ID,
		Title:   tr.Title,
		Content: tr.Content,
	}
	if tr.LinkedAttachments != nil {
		clone.LinkedAttachments = make([]*Attachment, len(tr.LinkedAttachments))
		copy(clone.LinkedAttachments, tr.LinkedAttachments)
	}
	if tr.Modified != nil {
		clonedModified := *tr.Modified
		clone.Modified = &clonedModified
	}
	return clone
}

// String returns a string representation of the trouble report.
func (tr *TroubleReport) String() string {
	return fmt.Sprintf("TroubleReport{ID: %d, Title: %s, Attachments: %d}",
		tr.ID, tr.Title, tr.AttachmentCount())
}

// Equals checks if two trouble reports are equal.
func (tr *TroubleReport) Equals(other *TroubleReport) bool {
	if other == nil {
		return false
	}
	if tr.ID != other.ID || tr.Title != other.Title || tr.Content != other.Content {
		return false
	}
	if tr.AttachmentCount() != other.AttachmentCount() {
		return false
	}
	for i, attachment := range tr.LinkedAttachments {
		if attachment != other.LinkedAttachments[i] {
			return false
		}
	}
	return true
}
