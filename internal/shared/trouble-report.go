package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

type TroubleReport struct {
	ID                EntityID `json:"id"`
	Title             string   `json:"title"`
	Content           string   `json:"content"`
	LinkedAttachments []string `json:"linked_attachments"` // LinkedAttachments is a list with paths to the images
	UseMarkdown       bool     `json:"use_markdown"`
}

func (tr *TroubleReport) Validate() *errors.ValidationError {
	if tr.Title == "" {
		return errors.NewValidationError("title cannot be empty")
	}

	if tr.Content == "" {
		return errors.NewValidationError("content cannot be empty")
	}

	return nil
}

func (tr *TroubleReport) Clone() *TroubleReport {
	clone := *tr
	clone.LinkedAttachments = make([]string, len(tr.LinkedAttachments))
	copy(clone.LinkedAttachments, tr.LinkedAttachments)
	return &clone
}

func (tr *TroubleReport) String() string {
	return fmt.Sprintf(
		"{ID: %d, Title: %s, Attachments: %d, UseMarkdown: %t}",
		tr.ID, tr.Title, len(tr.LinkedAttachments), tr.UseMarkdown,
	)
}

func (tr *TroubleReport) German() string {
	return tr.Title
}
