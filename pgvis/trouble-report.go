// Package pgvis trouble report models.
//
// This file defines the TroubleReport data structure and its associated
// validation and utility methods. Trouble reports are used to track
// issues, problems, and their resolutions within the system.
package pgvis

import (
	"strings"
)

const (
	// Validation constants for trouble reports
	MinTitleLength   = 1
	MaxTitleLength   = 500
	MinContentLength = 1
	MaxContentLength = 50000
)

// TroubleReport represents a trouble report in the system.
// It contains information about an issue, problem, or incident
// that needs to be tracked and potentially resolved.
type TroubleReport struct {
	// ID is the unique identifier for the trouble report
	ID int `json:"id"`

	// Title provides a brief summary of the trouble report
	Title string `json:"title"`

	// Content contains the detailed description of the issue
	Content string `json:"content"`

	// LinkedAttachments contains any files or documents related to this report
	LinkedAttachments []*Attachment `json:"linked_attachments"`

	// Modified tracks modification history and metadata
	Modified *Modified[*TroubleReport] `json:"modified"`
}

// NewTroubleReport creates a new trouble report with the provided details.
// It initializes an empty attachments slice and validates the input data.
//
// Parameters:
//   - m: Modification metadata for tracking changes
//   - title: Brief summary of the trouble report
//   - content: Detailed description of the issue
//
// Returns:
//   - *TroubleReport: The newly created trouble report
func NewTroubleReport(m *Modified[*TroubleReport], title, content string) *TroubleReport {
	return &TroubleReport{
		Title:             strings.TrimSpace(title),
		Content:           strings.TrimSpace(content),
		LinkedAttachments: make([]*Attachment, 0),
		Modified:          m,
	}
}

// NewBasicTroubleReport creates a trouble report with minimal metadata.
// This is useful for simple cases where modification tracking is not needed.
//
// Parameters:
//   - title: Brief summary of the trouble report
//   - content: Detailed description of the issue
//
// Returns:
//   - *TroubleReport: The newly created trouble report
func NewBasicTroubleReport(title, content string) *TroubleReport {
	systemUser := &User{
		UserName: "system",
	}
	modified := NewModified(systemUser, (*TroubleReport)(nil))

	return NewTroubleReport(modified, title, content)
}

// Validate checks if the trouble report has valid data.
//
// Returns:
//   - error: ValidationError for the first validation failure, or nil if valid
func (tr *TroubleReport) Validate() error {
	// Validate title
	if tr.Title == "" {
		return NewValidationError("title", "cannot be empty", tr.Title)
	}
	if len(tr.Title) < MinTitleLength {
		return NewValidationError("title", "too short", len(tr.Title))
	}
	if len(tr.Title) > MaxTitleLength {
		return NewValidationError("title", "too long", len(tr.Title))
	}

	// Validate content
	if tr.Content == "" {
		return NewValidationError("content", "cannot be empty", tr.Content)
	}
	if len(tr.Content) < MinContentLength {
		return NewValidationError("content", "too short", len(tr.Content))
	}
	if len(tr.Content) > MaxContentLength {
		return NewValidationError("content", "too long", len(tr.Content))
	}

	// Validate attachments if present
	if tr.LinkedAttachments != nil {
		for i, attachment := range tr.LinkedAttachments {
			if attachment == nil {
				return NewValidationError("linked_attachments",
					"attachment cannot be nil", i)
			}
		}
	}

	return nil
}

// UpdateTitle updates the trouble report title with validation.
//
// Parameters:
//   - newTitle: The new title for the trouble report
//
// Returns:
//   - error: Validation error if the title is invalid
func (tr *TroubleReport) UpdateTitle(newTitle string) error {
	newTitle = strings.TrimSpace(newTitle)

	if newTitle == "" {
		return NewValidationError("title", "cannot be empty", newTitle)
	}

	if len(newTitle) < MinTitleLength {
		return NewValidationError("title", "too short", len(newTitle))
	}

	if len(newTitle) > MaxTitleLength {
		return NewValidationError("title", "too long", len(newTitle))
	}

	tr.Title = newTitle
	return nil
}

// UpdateContent updates the trouble report content with validation.
//
// Parameters:
//   - newContent: The new content for the trouble report
//
// Returns:
//   - error: Validation error if the content is invalid
func (tr *TroubleReport) UpdateContent(newContent string) error {
	newContent = strings.TrimSpace(newContent)

	if newContent == "" {
		return NewValidationError("content", "cannot be empty", newContent)
	}

	if len(newContent) < MinContentLength {
		return NewValidationError("content", "too short", len(newContent))
	}

	if len(newContent) > MaxContentLength {
		return NewValidationError("content", "too long", len(newContent))
	}

	tr.Content = newContent
	return nil
}

// AddAttachment adds a new attachment to the trouble report.
//
// Parameters:
//   - attachment: The attachment to add
//
// Returns:
//   - error: Validation error if the attachment is invalid
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
//
// Parameters:
//   - index: The index of the attachment to remove
//
// Returns:
//   - error: Validation error if the index is invalid
func (tr *TroubleReport) RemoveAttachment(index int) error {
	if len(tr.LinkedAttachments) == 0 {
		return NewValidationError("index", "no attachments to remove", index)
	}

	if index < 0 || index >= len(tr.LinkedAttachments) {
		return NewValidationError("index", "out of range", index)
	}

	// Remove the attachment at the specified index
	tr.LinkedAttachments = append(tr.LinkedAttachments[:index], tr.LinkedAttachments[index+1:]...)
	return nil
}

// HasAttachments returns true if the trouble report has any attachments.
func (tr *TroubleReport) HasAttachments() bool {
	return len(tr.LinkedAttachments) > 0
}

// AttachmentCount returns the number of attachments.
func (tr *TroubleReport) AttachmentCount() int {
	if tr.LinkedAttachments == nil {
		return 0
	}
	return len(tr.LinkedAttachments)
}

// GetSummary returns a brief summary of the trouble report for display purposes.
//
// Parameters:
//   - maxLength: Maximum length of the summary
//
// Returns:
//   - string: Truncated content if longer than maxLength
func (tr *TroubleReport) GetSummary(maxLength int) string {
	if maxLength <= 0 {
		return ""
	}

	content := strings.TrimSpace(tr.Content)
	if len(content) <= maxLength {
		return content
	}

	// Truncate and add ellipsis
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

	// Clone attachments
	if tr.LinkedAttachments != nil {
		clone.LinkedAttachments = make([]*Attachment, len(tr.LinkedAttachments))
		copy(clone.LinkedAttachments, tr.LinkedAttachments)
	}

	// Clone modification data (shallow copy is sufficient for this use case)
	if tr.Modified != nil {
		clonedModified := *tr.Modified
		clone.Modified = &clonedModified
	}

	return clone
}

// String returns a string representation of the trouble report.
func (tr *TroubleReport) String() string {
	return "TroubleReport{ID: " + string(rune(tr.ID)) +
		", Title: " + tr.Title +
		", Attachments: " + string(rune(tr.AttachmentCount())) + "}"
}

// Equals checks if two trouble reports are equal.
func (tr *TroubleReport) Equals(other *TroubleReport) bool {
	if other == nil {
		return false
	}

	if tr.ID != other.ID ||
		tr.Title != other.Title ||
		tr.Content != other.Content {
		return false
	}

	// Compare attachments count
	if tr.AttachmentCount() != other.AttachmentCount() {
		return false
	}

	// Compare individual attachments (simplified comparison)
	for i, attachment := range tr.LinkedAttachments {
		if attachment != other.LinkedAttachments[i] {
			return false
		}
	}

	return true
}
