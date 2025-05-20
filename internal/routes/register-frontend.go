// Routes:
//
//   - GET: "/"
//   - GET: "/metal-sheets?tableSearch=..."
package routes

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

func RegisterPages(e *echo.Echo, o *Options) {
	e.GET(o.Global.ServerPathPrefix+"/", func(c echo.Context) error {
		err := serveTemplate(
			c, o.Templates, o.Global,
			"main.go.html",
			"layout/base.go.html",
			"content/home.go.html",
		)
		if err != nil {
			slog.Error(err.Error())
		}
		return err
	})

	e.GET(o.Global.ServerPathPrefix+"/metal-sheets", func(c echo.Context) error {
		// TODO: Handle query "tableSearch", pass to `o.MetalSheets()`
		//
		// tableSearch := c.QueryParam("tableSearch")
		slog.Debug("query", "QueryParams", c.QueryParams())

		err := serveTemplate(
			c, o.Templates, o.MetalSheets(),
			"main.go.html",
			"layout/base.go.html",
			"content/metal-sheets.tmpl",
		)
		if err != nil {
			slog.Error(err.Error())
		}
		return err
	})
}
