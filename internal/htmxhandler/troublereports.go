package htmxhandler

import "github.com/labstack/echo/v4"

type TroubleReports struct {
	*Base
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	// Dialog edit routes
	editDialogPath := h.ServerPathPrefix + "/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)

	// Data routes
	dataPath := h.ServerPathPrefix + "/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	attachmentsPreviewPath := h.ServerPathPrefix + "/attachments-preview"
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)

	sharePdfPath := h.ServerPathPrefix + "/share-pdf"
	e.GET(sharePdfPath, h.handleGetSharePdf)

	// Modifications routes
	modificationsPath := h.ServerPathPrefix + "/modifications/:id"
	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})
	e.POST(modificationsPath, h.handlePostModifications)

	// Attachment routes
	attachmentPath := h.ServerPathPrefix + "/attachments"
	e.GET(attachmentPath, h.handleGetAttachment)
}
