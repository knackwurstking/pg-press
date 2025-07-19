// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
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
}

func handleMainPage(templates fs.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.TroubleReportsTemplatePath,
			shared.NavFeedTemplatePath,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				pgvis.WrapError(err, "failed to parse templates"))
		}

		if err = t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				pgvis.WrapError(err, "failed to render page"))
		}

		return nil
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
