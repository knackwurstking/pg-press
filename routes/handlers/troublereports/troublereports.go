// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
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

	// TODO: ...
	editDialogPath := h.serverPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, handleGetEditDialog(h.templates, h.db))
	e.POST(editDialogPath, handleCreateReport(h.templates, h.db))
	e.PUT(editDialogPath, handleUpdateReport(h.templates, h.db))

	// TODO: ...
	dataPath := h.serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, handleGetData(h.templates, h.db))
	e.DELETE(dataPath, handleDeleteReport(h.templates, h.db))

	// TODO: ...
	modificationsPath := h.serverPathPrefix + "/trouble-reports/modifications"
	e.GET(modificationsPath, handleGetModifications(h.templates, h.db))
}

func (h *Handler) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.templates,
		constants.TroubleReportsPageTemplates,
	)
}

func handleGetEditDialog(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETDialogEdit(templates, c, db, nil)
	}
}

func handleCreateReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return POSTDialogEdit(templates, c, db)
	}
}

func handleUpdateReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return PUTDialogEdit(templates, c, db)
	}
}

func handleGetData(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETData(templates, c, db)
	}
}

func handleDeleteReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return DELETEData(templates, c, db)
	}
}

func handleGetModifications(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETModifications(templates, c, db)
	}
}
