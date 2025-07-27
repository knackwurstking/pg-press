package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

type TroubleReports struct {
	*Base
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/trouble-reports", h.handleMainPage)

	// Dialog edit routes
	editDialogPath := h.ServerPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)

	// Data routes
	dataPath := h.ServerPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	attachmentsPreviewPath := h.ServerPathPrefix + "/trouble-reports/attachments-preview"
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)

	sharePdfPath := h.ServerPathPrefix + "/trouble-reports/share-pdf"
	e.GET(sharePdfPath, h.handleGetSharePdf)

	// Modifications routes
	modificationsPath := h.ServerPathPrefix + "/trouble-reports/modifications/:id"
	modificationsAttachmentsPath := h.ServerPathPrefix +
		"/trouble-reports/modifications/attachments-preview/:id"
	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})
	e.GET(modificationsAttachmentsPath, h.handleGetModificationAttachmentsPreview)
	e.POST(modificationsPath, h.handlePostModifications)

	// Attachment routes
	attachmentReorderPath := h.ServerPathPrefix + "/trouble-reports/attachments/reorder"
	e.POST(attachmentReorderPath, h.handlePostAttachmentReorder)

	attachmentPath := h.ServerPathPrefix + "/trouble-reports/attachments"
	e.GET(attachmentPath, h.handleGetAttachment)
	e.DELETE(attachmentPath, h.handleDeleteAttachment)
}

func (h *TroubleReports) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.Templates,
		constants.TroubleReportsPageTemplates,
	)
}
