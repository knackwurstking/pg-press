package htmxhandler

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
type troubleReportsDataTemplateData struct {
	TroubleReports []*database.TroubleReportWithAttachments `json:"trouble_reports"`
	User           *database.User                           `json:"user"`
}

type dialogEditTemplateData struct {
	ID                int                    `json:"id"`
	Submitted         bool                   `json:"submitted"`
	Title             string                 `json:"title"`
	Content           string                 `json:"content"`
	LinkedAttachments []*database.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                   `json:"invalid_title"`
	InvalidContent    bool                   `json:"invalid_content"`
	AttachmentError   string                 `json:"attachment_error,omitempty"`
}

type attachmentsPreviewTemplateData struct {
	TroubleReport *database.TroubleReportWithAttachments `json:"trouble_report"`
}

type modificationsTemplateData struct {
	User          *database.User
	TroubleReport *database.TroubleReport
	Mods          database.Mods[database.TroubleReportMod]
}

func (mtd *modificationsTemplateData) FirstModified() *database.Modified[database.TroubleReportMod] {
	if len(mtd.TroubleReport.Mods) == 0 {
		return nil
	}
	return mtd.TroubleReport.Mods[0]
}

type modificationAttachmentsTemplateData struct {
	TroubleReport *database.TroubleReport
	Modification  *database.Modified[database.TroubleReportMod]
	Attachments   []*database.Attachment
}
