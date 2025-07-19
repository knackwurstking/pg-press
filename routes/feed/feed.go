// Package feed provides HTTP route handlers for feed management.
package feed

import (
	"io/fs"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// Serve configures and registers all feed related HTTP routes.
func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/feed", handleMainPage(templates))
	e.GET(serverPathPrefix+"/feed/data", handleGetData(templates, db))
}

// handleMainPage returns a handler for the main feed page.
func handleMainPage(templates fs.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		return utils.HandleTemplate(c, nil,
			templates,
			[]string{
				shared.LayoutTemplatePath,
				shared.FeedTemplatePath,
			},
		)
	}
}

// handleGetData returns a handler for the feed data endpoint.
func handleGetData(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETData(templates, c, db)
	}
}
