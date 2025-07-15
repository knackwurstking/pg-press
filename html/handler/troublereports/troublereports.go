package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/trouble-reports", func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/trouble-reports.html",
			"templates/nav/feed.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err = t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	// HTMX: Dialog Edit

	e.GET(serverPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return GETDialogEdit(templates, c, db, nil)
	})

	// FormValues:
	//   - title: string
	//   - content: multiline-string
	e.POST(serverPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return POSTDialogEdit(templates, c, db)
	})

	// QueryParam:
	//   - id: int
	//
	// FormValue:
	//   - title: string
	//   - content: multiline-string
	e.PUT(serverPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return PUTDialogEdit(templates, c, db)
	})

	// HTMX: Data

	e.GET(serverPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return GETData(templates, c, db)
	})

	// QueryParam:
	//   - id: int
	e.DELETE(serverPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return DELETEData(templates, c, db)
	})
}
