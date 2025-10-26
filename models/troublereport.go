// TODO: Remove useless stuff
package models

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/errors"
)

const (
	MinTitleLength   = 1
	MaxTitleLength   = 500
	MinContentLength = 1
	MaxContentLength = 50000
)

type TroubleReportID int64

// TroubleReport represents a trouble report in the system.
type TroubleReport struct {
	ID                TroubleReportID `json:"id"`
	Title             string          `json:"title"`
	Content           string          `json:"content"`
	LinkedAttachments []AttachmentID  `json:"linked_attachments"`
	UseMarkdown       bool            `json:"use_markdown"`
}

// New creates a new trouble report with the provided details.
func NewTroubleReport(title, content string, linkedAttachments ...AttachmentID) *TroubleReport {
	return &TroubleReport{
		Title:             strings.TrimSpace(title),
		Content:           strings.TrimSpace(content),
		LinkedAttachments: linkedAttachments,
		UseMarkdown:       false,
	}
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
		return errors.NewValidationError("cannot be empty")
	}
	if len(title) < MinTitleLength {
		return errors.NewValidationError("title too short")
	}
	if len(title) > MaxTitleLength {
		return errors.NewValidationError("title too long")
	}
	return nil
}

func (tr *TroubleReport) validateContent(content string) error {
	if content == "" {
		return errors.NewValidationError("content cannot be empty")
	}
	if len(content) < MinContentLength {
		return errors.NewValidationError("content too short")
	}
	if len(content) > MaxContentLength {
		return errors.NewValidationError("content too long")
	}
	return nil
}

func (tr *TroubleReport) validateAttachments() error {
	for _, attachmentID := range tr.LinkedAttachments {
		if attachmentID <= 0 {
			return errors.NewValidationError("linked_attachments: attachment ID must be positive")
		}
	}
	return nil
}

// AddAttachment adds a new attachment ID to the trouble report.
func (tr *TroubleReport) AddAttachment(attachmentID AttachmentID) error {
	if attachmentID <= 0 {
		return errors.NewValidationError("linked_attachments: attachment ID must be positive")
	}
	if tr.LinkedAttachments == nil {
		tr.LinkedAttachments = make([]AttachmentID, 0)
	}
	tr.LinkedAttachments = append(tr.LinkedAttachments, attachmentID)
	return nil
}

// RemoveAttachment removes an attachment by index.
func (tr *TroubleReport) RemoveAttachment(index int) error {
	if len(tr.LinkedAttachments) == 0 {
		return errors.NewValidationError("index: no attachments to remove")
	}

	if index < 0 || index >= len(tr.LinkedAttachments) {
		return errors.NewValidationError("index: out of range")
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
	if tr.ID != other.ID || tr.Title != other.Title || tr.Content != other.Content || tr.UseMarkdown != other.UseMarkdown {
		return false
	}
	if tr.AttachmentCount() != other.AttachmentCount() {
		return false
	}
	for i, attachmentID := range tr.LinkedAttachments {
		if attachmentID != other.LinkedAttachments[i] {
			return false
		}
	}
	return true
}

// TroubleReportWithAttachments represents a trouble report with its attachments loaded.
type TroubleReportWithAttachments struct {
	*TroubleReport
	LoadedAttachments []*Attachment `json:"loaded_attachments"`
}
