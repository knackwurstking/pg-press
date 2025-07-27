package handler

import (
	"github.com/knackwurstking/pg-vis/internal/database"
)

const (
	adminPrivilegesRequiredMessage   = "administrator privileges required"
	invalidContentFormFieldMessage   = "invalid content form value"
	invalidTitleFormFieldMessage     = "invalid title form value"
	attachmentTooLargeMessage        = "image exceeds maximum size limit (10MB)"
	attachmentNotFoundMessage        = "image not found"
	invalidAttachmentMessage         = "invalid image data"
	nonImageFileMessage              = "only image files are allowed (JPG, PNG, GIF, BMP, SVG, WebP)"
	tooManyAttachmentsMessage        = "too many images (maximum 10 allowed)"
	attachmentProcessingErrorMessage = "failed to process image"
)

// Template data structures
type TroubleReportsDataTemplateData struct {
	TroubleReports []*database.TroubleReportWithAttachments `json:"trouble_reports"`
	User           *database.User                           `json:"user"`
}

type DialogEditTemplateData struct {
	ID                int                    `json:"id"`
	Submitted         bool                   `json:"submitted"`
	Title             string                 `json:"title"`
	Content           string                 `json:"content"`
	LinkedAttachments []*database.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                   `json:"invalid_title"`
	InvalidContent    bool                   `json:"invalid_content"`
	AttachmentError   string                 `json:"attachment_error,omitempty"`
}

type AttachmentsPreviewTemplateData struct {
	TroubleReport *database.TroubleReportWithAttachments `json:"trouble_report"`
}

type ModificationsTemplateData struct {
	User              *database.User
	TroubleReport     *database.TroubleReport
	LoadedAttachments []*database.Attachment
	Mods              database.Mods[database.TroubleReportMod]
}

func (mtd *ModificationsTemplateData) FirstModified() *database.Modified[database.TroubleReportMod] {
	if len(mtd.TroubleReport.Mods) == 0 {
		return nil
	}
	return mtd.TroubleReport.Mods[0]
}

type ModificationAttachmentsTemplateData struct {
	TroubleReport *database.TroubleReport
	Modification  *database.Modified[database.TroubleReportMod]
	Attachments   []*database.Attachment
}
