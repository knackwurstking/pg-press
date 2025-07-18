// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// DataPageData contains the data structure passed to the trouble reports data template.
// This structure includes all trouble reports and user permission information needed
// for rendering the data table and control elements.
type DataPageData struct {
	// TroubleReports contains all trouble reports to display in the data table
	TroubleReports []*pgvis.TroubleReport `json:"trouble_reports"`

	// IsAdmin indicates whether the current user has administrator privileges
	// This controls the visibility of administrative actions like delete buttons
	IsAdmin bool `json:"is_admin"`
}

// GETData handles GET requests to retrieve and render trouble reports data.
// This function is typically called via HTMX to dynamically load or refresh
// the trouble reports data table without requiring a full page reload.
//
// The function performs the following operations:
//  1. Authenticates the user and retrieves user context
//  2. Fetches all trouble reports from the database
//  3. Determines user permissions for administrative functions
//  4. Renders the data template with the collected information
//
// Parameters:
//   - templates: Embedded filesystem containing HTML templates
//   - c: Echo context containing request/response information
//   - db: Database connection for data operations
//
// Returns:
//   - *echo.HTTPError: HTTP error if operation fails, nil on success
//
// HTTP Response:
//   - 200 OK: Successfully rendered trouble reports data
//   - 401 Unauthorized: User authentication failed
//   - 500 Internal Server Error: Database or template errors
func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	// Authenticate user and retrieve context information
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Fetch all trouble reports from database
	trs, err := db.TroubleReports.List()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Parse the data template
	t, err := template.ParseFS(templates, shared.TroubleReportsDataTemplatePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			pgvis.WrapError(err, "failed to load page template"))
	}

	// Prepare page data with trouble reports and user permissions
	pageData := DataPageData{
		TroubleReports: trs,
		IsAdmin:        user.IsAdmin(),
	}

	// Execute template with the prepared data
	if err = t.Execute(c.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			pgvis.WrapError(err, "Failed to render page"))
	}

	return nil
}

// DELETEData handles DELETE requests to remove trouble reports.
// This function implements administrative deletion of trouble reports with
// proper authorization checks and audit logging.
//
// The function performs the following operations:
//  1. Validates and parses the report ID from query parameters
//  2. Authenticates the user and checks administrative privileges
//  3. Logs the deletion attempt for audit purposes
//  4. Removes the trouble report from the database
//  5. Returns updated data table via GETData
//
// Required Query Parameters:
//   - id: int (positive integer identifying the report to delete)
//
// Parameters:
//   - templates: Embedded filesystem containing HTML templates
//   - c: Echo context containing request/response information
//   - db: Database connection for data operations
//
// Returns:
//   - *echo.HTTPError: HTTP error if operation fails, nil on success
//
// HTTP Response:
//   - 200 OK: Successfully deleted report and returned updated data
//   - 400 Bad Request: Invalid or missing ID parameter
//   - 401 Unauthorized: User authentication failed
//   - 403 Forbidden: User lacks administrator privileges
//   - 404 Not Found: Report with specified ID not found
//   - 500 Internal Server Error: Database or template errors
//
// Security Notes:
//   - Requires administrator privileges
//   - All deletion attempts are logged for audit purposes
//   - TODO: Implement voting system for non-administrator deletions
func DELETEData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	// Parse and validate the report ID from query parameters
	id, herr := utils.ParseRequiredIDQuery(c, "id")
	if herr != nil {
		return herr
	}

	// Authenticate user and retrieve context information
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Check administrator privileges
	if !user.IsAdmin() {
		log.Warnf("Non-admin user %s (ID: %d) attempted to delete trouble report %d",
			user.UserName, user.TelegramID, id)

		// TODO: Implement voting system for deletion by non-administrators
		// For now, only administrators can delete trouble reports
		log.Warnf("Voting system for deletion not implemented. Administrator privileges required.")

		return echo.NewHTTPError(http.StatusForbidden,
			"Administrator privileges required for deletion")
	}

	// Log the deletion attempt for audit purposes
	log.Infof("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	// Attempt to remove the trouble report from the database
	if err := db.TroubleReports.Remove(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return updated data table by calling GETData
	// This ensures the client receives the current state after deletion
	return GETData(templates, c, db)
}
