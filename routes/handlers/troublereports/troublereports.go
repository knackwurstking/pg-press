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

	dialogEditHandler    *DialogEditHandler
	dataHandler          *DataHandler
	modificationsHandler *ModificationsHandler
	attachmentsHandler   *AttachmentsHandler
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS) *Handler {
	dialogEditHandler := &DialogEditHandler{db, serverPathPrefix, templates}

	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,

		dialogEditHandler:    dialogEditHandler,
		dataHandler:          &DataHandler{db, serverPathPrefix, templates},
		modificationsHandler: &ModificationsHandler{db, serverPathPrefix, templates},

		attachmentsHandler: &AttachmentsHandler{
			db:                db,
			serverPathPrefix:  serverPathPrefix,
			templates:         templates,
			dialogEditHandler: dialogEditHandler,
		},
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/trouble-reports", h.handleMainPage)

	h.dialogEditHandler.RegisterRoutes(e)

	h.dataHandler.RegisterRoutes(e)

	h.modificationsHandler.RegisterRoutes(e)

	// Attachment management routes
	h.attachmentsHandler.RegisterRoutes(e)
}

func (h *Handler) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.templates,
		constants.TroubleReportsPageTemplates,
	)
}
