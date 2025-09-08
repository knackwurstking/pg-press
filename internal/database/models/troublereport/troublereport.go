package troublereport

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models/attachment"
	"github.com/knackwurstking/pgpress/internal/database/models/mod"
)

const (
	MinTitleLength   = 1
	MaxTitleLength   = 500
	MinContentLength = 1
	MaxContentLength = 50000
)

type TroubleReportMod struct {
	Title             string
	Content           string
	LinkedAttachments []int64
}

// TroubleReport represents a trouble report in the system.
type TroubleReport struct {
	ID                int64                      `json:"id"`
	Title             string                     `json:"title"`
	Content           string                     `json:"content"`
	LinkedAttachments []int64                    `json:"linked_attachments"`
	Mods              mod.Mods[TroubleReportMod] `json:"mods"`
}

// NewTroubleReport creates a new trouble report with the provided details.
func NewTroubleReport(title, content string) *TroubleReport {
	return &TroubleReport{
		Title:             strings.TrimSpace(title),
		Content:           strings.TrimSpace(content),
		LinkedAttachments: make([]int64, 0),
		Mods:              mod.NewMods[TroubleReportMod](),
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
		return dberror.NewValidationError("title", "cannot be empty", title)
	}
	if len(title) < MinTitleLength {
		return dberror.NewValidationError("title", "too short", len(title))
	}
	if len(title) > MaxTitleLength {
		return dberror.NewValidationError("title", "too long", len(title))
	}
	return nil
}

func (tr *TroubleReport) validateContent(content string) error {
	if content == "" {
		return dberror.NewValidationError("content", "cannot be empty", content)
	}
	if len(content) < MinContentLength {
		return dberror.NewValidationError("content", "too short", len(content))
	}
	if len(content) > MaxContentLength {
		return dberror.NewValidationError("content", "too long", len(content))
	}
	return nil
}

func (tr *TroubleReport) validateAttachments() error {
	for i, attachmentID := range tr.LinkedAttachments {
		if attachmentID <= 0 {
			return dberror.NewValidationError("linked_attachments", "attachment ID must be positive", i)
		}
	}
	return nil
}

// AddAttachment adds a new attachment ID to the trouble report.
func (tr *TroubleReport) AddAttachment(attachmentID int64) error {
	if attachmentID <= 0 {
		return dberror.NewValidationError("attachment_id", "must be positive", attachmentID)
	}
	if tr.LinkedAttachments == nil {
		tr.LinkedAttachments = make([]int64, 0)
	}
	tr.LinkedAttachments = append(tr.LinkedAttachments, attachmentID)
	return nil
}

// RemoveAttachment removes an attachment by index.
func (tr *TroubleReport) RemoveAttachment(index int) error {
	if len(tr.LinkedAttachments) == 0 {
		return dberror.NewValidationError("index", "no attachments to remove", index)
	}
	if index < 0 || index >= len(tr.LinkedAttachments) {
		return dberror.NewValidationError("index", "out of range", index)
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
	if tr.ID != other.ID || tr.Title != other.Title || tr.Content != other.Content {
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
	LoadedAttachments []*attachment.Attachment `json:"loaded_attachments"`
}
