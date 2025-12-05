package models

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pg-press/env"
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
	if tr.Title == "" {
		return fmt.Errorf("cannot be empty")
	}
	if len(tr.Title) < env.MinTitleLength {
		return fmt.Errorf("title too short")
	}
	if len(tr.Title) > env.MaxTitleLength {
		return fmt.Errorf("title too long")
	}

	if tr.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	if len(tr.Content) < env.MinContentLength {
		return fmt.Errorf("content too short")
	}
	if len(tr.Content) > env.MaxContentLength {
		return fmt.Errorf("content too long")
	}

	for _, attachmentID := range tr.LinkedAttachments {
		if attachmentID <= 0 {
			return fmt.Errorf("linked_attachments: attachment ID must be positive")
		}
	}
	return nil
}

// String returns a string representation of the trouble report.
func (tr *TroubleReport) String() string {
	return fmt.Sprintf("TroubleReport{ID: %d, Title: %s, Attachments: %d}",
		tr.ID, tr.Title, len(tr.LinkedAttachments))
}

// TroubleReportWithAttachments represents a trouble report with its attachments loaded.
type TroubleReportWithAttachments struct {
	*TroubleReport
	LoadedAttachments []*Attachment `json:"loaded_attachments"`
}
