package htmxhandler

import (
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	// Dialog edit routes
	editDialogPath := serverPathPrefix + "/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)

	// Data routes
	dataPath := serverPathPrefix + "/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	attachmentsPreviewPath := serverPathPrefix + "/attachments-preview"
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)

	sharePdfPath := serverPathPrefix + "/share-pdf"
	e.GET(sharePdfPath, h.handleGetSharePdf)

	// Modifications routes
	modificationsPath := serverPathPrefix + "/modifications/:id"
	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})
	e.POST(modificationsPath, h.handlePostModifications)

	// Attachment routes
	attachmentPath := serverPathPrefix + "/attachments"
	e.GET(attachmentPath, h.handleGetAttachment)
}
