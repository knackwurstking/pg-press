package htmxhandler

import (
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	logger.HTMXHandlerTroubleReports().Debug("Registering trouble reports HTMX routes")

	// Dialog edit routes
	editDialogPath := serverPathPrefix + "/htmx/trouble-reports/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.GET(editDialogPath+"/", func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.POST(editDialogPath+"/", h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)
	e.PUT(editDialogPath+"/", h.handlePutDialogEdit)

	// Data routes
	dataPath := serverPathPrefix + "/htmx/trouble-reports/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	attachmentsPreviewPath := serverPathPrefix + "/htmx/trouble-reports/attachments-preview"
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)

	// Modifications routes
	modificationsPath := serverPathPrefix + "/htmx/trouble-reports/modifications/:id"
	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})
	e.POST(modificationsPath, h.handlePostModifications)
}
