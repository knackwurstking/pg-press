// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

const (
	adminPrivilegesRequiredMessage   = "administrator privileges required"
	invalidContentFormFieldMessage   = "invalid content form value"
	invalidTitleFormFieldMessage     = "invalid title form value"
	attachmentTooLargeMessage        = "attachment exceeds maximum size limit (10MB)"
	attachmentNotFoundMessage        = "attachment not found"
	invalidAttachmentMessage         = "invalid attachment data"
	tooManyAttachmentsMessage        = "too many attachments (maximum 10 allowed)"
	attachmentProcessingErrorMessage = "failed to process attachment"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/trouble-reports", h.handleMainPage)

	editDialogPath := h.serverPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)

	dataPath := h.serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	modificationsPath := h.serverPathPrefix + "/trouble-reports/modifications/:id"
	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})
	e.POST(modificationsPath, h.handlePostModifications)

	// Attachment management routes
	attachmentReorderPath := h.serverPathPrefix + "/trouble-reports/attachments/reorder"
	e.POST(attachmentReorderPath, h.handlePostAttachmentReorder)

	attachmentPath := h.serverPathPrefix + "/trouble-reports/attachments"
	e.GET(attachmentPath, h.handleGetAttachment)
	e.DELETE(attachmentPath, h.handleDeleteAttachment)
}

func (h *Handler) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.templates,
		constants.TroubleReportsPageTemplates,
	)
}
