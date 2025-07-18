// Package troublereports provides HTTP route handlers for trouble report management.
//
// This package implements the web interface for creating, viewing, editing, and deleting
// trouble reports. It handles both the main page rendering and HTMX-based interactions
// for dynamic updates without full page reloads.
//
// Key Features:
//   - Main trouble reports page with data listing
//   - Dynamic edit dialog for creating/updating reports
//   - Admin-only deletion functionality
//   - Form validation and sanitization
//   - Integration with the pgvis database layer
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
//
// This function sets up the following endpoints:
//   - GET  /trouble-reports           - Main trouble reports page
//   - GET  /trouble-reports/dialog-edit - Edit dialog (HTMX)
//   - POST /trouble-reports/dialog-edit - Create new report
//   - PUT  /trouble-reports/dialog-edit - Update existing report
//   - GET  /trouble-reports/data       - Data listing (HTMX)
//   - DELETE /trouble-reports/data     - Delete report (admin only)
//
// Parameters:
//   - templates: Embedded filesystem containing HTML templates
//   - serverPathPrefix: URL prefix for all routes (e.g., "/app")
//   - e: Echo instance to register routes with
//   - db: Database connection for data operations
func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	// Main trouble reports page
	e.GET(serverPathPrefix+"/trouble-reports", handleMainPage(templates))

	// Edit dialog endpoints (HTMX)
	editDialogPath := serverPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, handleGetEditDialog(templates, db))
	e.POST(editDialogPath, handleCreateReport(templates, db))
	e.PUT(editDialogPath, handleUpdateReport(templates, db))

	// Data endpoints (HTMX)
	dataPath := serverPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, handleGetData(templates, db))
	e.DELETE(dataPath, handleDeleteReport(templates, db))
}

// handleMainPage returns a handler for the main trouble reports page.
// This page serves as the container for the trouble reports interface,
// loading the necessary templates and navigation components.
func handleMainPage(templates fs.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse required templates for the main page
		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.TroubleReportsTemplatePath,
			shared.NavFeedTemplatePath,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to parse templates: "+err.Error())
		}

		// Render the main page without any data
		if err = t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to render page: "+err.Error())
		}

		return nil
	}
}

// handleGetEditDialog returns a handler for the edit dialog GET requests.
// This endpoint is used by HTMX to load the edit dialog, either empty for
// new reports or populated with existing data for editing.
func handleGetEditDialog(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETDialogEdit(templates, c, db, nil)
	}
}

// handleCreateReport returns a handler for creating new trouble reports.
// This endpoint processes POST requests with form data to create new reports.
//
// Expected form values:
//   - title: string (required, 1-500 characters)
//   - content: string (required, 1-50000 characters)
func handleCreateReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return POSTDialogEdit(templates, c, db)
	}
}

// handleUpdateReport returns a handler for updating existing trouble reports.
// This endpoint processes PUT requests with form data and a query parameter
// specifying which report to update.
//
// Required query parameter:
//   - id: int (positive integer identifying the report)
//
// Expected form values:
//   - title: string (required, 1-500 characters)
//   - content: string (required, 1-50000 characters)
func handleUpdateReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return PUTDialogEdit(templates, c, db)
	}
}

// handleGetData returns a handler for the data listing endpoint.
// This endpoint is used by HTMX to load and refresh the trouble reports
// data table without a full page reload.
func handleGetData(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETData(templates, c, db)
	}
}

// handleDeleteReport returns a handler for deleting trouble reports.
// This endpoint requires administrator privileges and processes DELETE
// requests with a query parameter specifying which report to delete.
//
// Required query parameter:
//   - id: int (positive integer identifying the report to delete)
//
// Access Control:
//   - Requires administrator privileges
//   - TODO: Implement voting system for non-admin deletions
func handleDeleteReport(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return DELETEData(templates, c, db)
	}
}
