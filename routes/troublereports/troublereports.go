// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// Serve configures and registers all trouble report related HTTP routes.
func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/trouble-reports", handleMainPage(templates))

	editDialogPath := serverPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, handleGetEditDialog(templates, db))
	e.POST(editDialogPath, handleCreateReport(templates, db))
	e.PUT(editDialogPath, handleUpdateReport(templates, db))

	dataPath := serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, handleGetData(templates, db))
	e.DELETE(dataPath, handleDeleteReport(templates, db))

	modificationsPath := serverPathPrefix + "/trouble-reports/modifications"
	e.GET(modificationsPath, handleGetModifications(templates, db))
}

func handleMainPage(templates fs.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		return utils.HandleTemplate(c, nil,
			templates,
			[]string{
				shared.LayoutTemplatePath,
				shared.TroubleReportsTemplatePath,
				shared.NavFeedTemplatePath,
			},
		)
	}
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
