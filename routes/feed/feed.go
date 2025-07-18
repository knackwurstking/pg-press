// Package feed provides HTTP route handlers for feed management.
//
// This package implements the web interface for viewing feeds and feed data.
// It handles both the main feed page rendering and HTMX-based interactions
// for dynamic content loading without full page reloads.
//
// Key Features:
//   - Main feed page with data listing
//   - Dynamic data loading for feed content
//   - Integration with the pgvis database layer
//
// Future Enhancements:
//   - WebSocket handler for nav icon button count per user
//   - WebSocket handler for live feed updates
package feed

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
)

// Serve configures and registers all feed related HTTP routes.
//
// This function sets up the following endpoints:
//   - GET /feed      - Main feed page
//   - GET /feed/data - Feed data listing (HTMX)
//
// Parameters:
//   - templates: Embedded filesystem containing HTML templates
//   - serverPathPrefix: URL prefix for all routes (e.g., "/app")
//   - e: Echo instance to register routes with
//   - db: Database connection for data operations
func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	// Main feed page
	e.GET(serverPathPrefix+"/feed", handleMainPage(templates))

	// Data endpoint (HTMX)
	e.GET(serverPathPrefix+"/feed/data", handleGetData(templates, db))
}

// handleMainPage returns a handler for the main feed page.
// This page serves as the container for the feed interface,
// loading the necessary templates for the feed display.
func handleMainPage(templates fs.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse required templates for the main feed page
		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.FeedTemplatePath,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to parse templates: "+err.Error())
		}

		// Render the main feed page without any data
		if err = t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to render page: "+err.Error())
		}

		return nil
	}
}

// handleGetData returns a handler for the feed data endpoint.
// This endpoint is used by HTMX to load and refresh the feed data
// without requiring a full page reload.
func handleGetData(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETData(templates, c, db)
	}
}
