// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// EditDialogPageData contains the data structure passed to the edit dialog template.
// This structure includes all necessary information for rendering the edit form,
// handling validation errors, and managing dialog state.
type EditDialogPageData struct {
	// ID is the trouble report identifier for editing existing reports
	// Set to 0 for new report creation
	ID int `json:"id"`

	// Submitted indicates whether the form has been successfully submitted
	// When true, the dialog will be closed on the client side
	Submitted bool `json:"submitted"`

	// Title contains the trouble report title text
	Title string `json:"title"`

	// Content contains the trouble report content/description
	Content string `json:"content"`

	// LinkedAttachments contains any file attachments associated with the report
	// Currently used for display purposes in edit mode
	LinkedAttachments []*pgvis.Attachment `json:"linked_attachments,omitempty"`

	// InvalidTitle indicates validation failure for the title field
	InvalidTitle bool `json:"invalid_title"`

	// InvalidContent indicates validation failure for the content field
	InvalidContent bool `json:"invalid_content"`
}

// GETDialogEdit handles GET requests for the trouble report edit dialog.
// This function serves the edit dialog interface, either empty for creating
// new reports or populated with existing data for editing.
//
// The function supports the following query parameters:
//   - cancel: "true" - Closes the dialog without saving changes
//   - id: int - Loads existing trouble report data for editing
//
// Parameters:
//   - templates: Embedded filesystem containing HTML templates
//   - c: Echo context containing request/response information
//   - db: Database connection for data operations
//   - pageData: Pre-populated page data (used for validation errors)
//
// Returns:
//   - *echo.HTTPError: HTTP error if operation fails, nil on success
//
// HTTP Response:
//   - 200 OK: Successfully rendered edit dialog
//   - 400 Bad Request: Invalid ID parameter
//   - 404 Not Found: Trouble report with specified ID not found
//   - 500 Internal Server Error: Database or template errors
func GETDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB, pageData *EditDialogPageData) *echo.HTTPError {
	// Initialize page data if not provided (new request)
	if pageData == nil {
		pageData = &EditDialogPageData{
			Submitted: false,
		}
	}

	// Handle cancel request - close dialog without saving
	if c.QueryParam(shared.CancelQueryParam) == shared.TrueValue {
		log.Debug("Edit dialog cancelled by user")
		pageData.Submitted = true
	}

	// Load existing trouble report data if editing and no validation errors
	if !pageData.Submitted && !pageData.InvalidTitle && !pageData.InvalidContent {
		if idStr := c.QueryParam(shared.IDQueryParam); idStr != "" {
			// Parse the ID parameter for editing existing reports
			id, herr := utils.ParseRequiredIDQuery(c, shared.IDQueryParam)
			if herr != nil {
				log.Warnf("Invalid ID parameter for edit dialog: %v", herr)
				return herr
			}

			log.Debugf("Loading trouble report data for editing: ID=%d", id)
			pageData.ID = int(id)

			// Fetch the existing trouble report from database
			tr, err := db.TroubleReports.Get(id)
			if err != nil {
				log.Errorf("Failed to retrieve trouble report %d: %v", id, err)
				return utils.HandlePgvisError(c, err)
			}

			// Populate form fields with existing data
			pageData.Title = tr.Title
			pageData.Content = tr.Content
			pageData.LinkedAttachments = tr.LinkedAttachments

			log.Debugf("Loaded trouble report for editing: title=%q", tr.Title)
		}
	}

	// Parse and execute the edit dialog template
	t, err := template.ParseFS(templates, shared.TroubleReportsDialogTemplatePath)
	if err != nil {
		log.Errorf("Failed to parse edit dialog template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"Failed to load dialog template: "+err.Error())
	}

	// Render the dialog with the prepared data
	if err = t.Execute(c.Response(), pageData); err != nil {
		log.Errorf("Failed to execute edit dialog template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"Failed to render dialog: "+err.Error())
	}

	return nil
}

// POSTDialogEdit handles POST requests to create new trouble reports.
// This function processes form submissions for creating new trouble reports,
// including validation, sanitization, and database storage.
//
// Expected form fields:
//   - title: string (required, 1-500 characters)
//   - content: string (required, 1-50000 characters)
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
//   - 200 OK: Successfully created report or returned validation errors
//   - 400 Bad Request: Invalid form data or validation failures
//   - 401 Unauthorized: User authentication failed
//   - 500 Internal Server Error: Database or template errors
func POSTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	// Initialize dialog data for new report creation
	dialogEditData := &EditDialogPageData{
		Submitted: true, // Assume success initially
	}

	// Authenticate user and retrieve context information
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		log.Warnf("Failed to get user from context for trouble report creation: %v", herr)
		return herr
	}

	// Extract and validate form data
	title, content, herr := extractAndValidateFormData(c)
	if herr != nil {
		log.Warnf("Form validation failed for new trouble report: %v", herr)
		return herr
	}

	// Store form data for potential redisplay on validation errors
	dialogEditData.Title = title
	dialogEditData.Content = content

	// Perform field validation
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	// Create new trouble report if validation passes
	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		log.Infof("Creating new trouble report by user %s: title=%q", user.UserName, title)

		// Create modification tracking metadata
		modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
		tr := pgvis.NewTroubleReport(modified, title, content)

		// Save to database
		if err := db.TroubleReports.Add(tr); err != nil {
			log.Errorf("Failed to create trouble report: %v", err)
			return utils.HandlePgvisError(c, err)
		}

		log.Infof("Successfully created trouble report ID=%d", tr.ID)
	} else {
		// Validation failed - keep dialog open for corrections
		dialogEditData.Submitted = false
		log.Debugf("Validation failed for new trouble report: title_invalid=%v, content_invalid=%v",
			dialogEditData.InvalidTitle, dialogEditData.InvalidContent)
	}

	// Return the dialog with updated state
	return GETDialogEdit(templates, c, db, dialogEditData)
}

// PUTDialogEdit handles PUT requests to update existing trouble reports.
// This function processes form submissions for updating existing trouble reports,
// including validation, sanitization, and database updates with change tracking.
//
// Required query parameter:
//   - id: int (positive integer identifying the report to update)
//
// Expected form fields:
//   - title: string (required, 1-500 characters)
//   - content: string (required, 1-50000 characters)
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
//   - 200 OK: Successfully updated report or returned validation errors
//   - 400 Bad Request: Invalid form data, ID parameter, or validation failures
//   - 401 Unauthorized: User authentication failed
//   - 404 Not Found: Trouble report with specified ID not found
//   - 500 Internal Server Error: Database or template errors
func PUTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	// Parse and validate the report ID
	id, herr := utils.ParseRequiredIDQuery(c, shared.IDQueryParam)
	if herr != nil {
		log.Warnf("Invalid ID parameter for trouble report update: %v", herr)
		return herr
	}

	// Authenticate user and retrieve context information
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		log.Warnf("Failed to get user from context for trouble report update: %v", herr)
		return herr
	}

	// Extract and validate form data
	title, content, herr := extractAndValidateFormData(c)
	if herr != nil {
		log.Warnf("Form validation failed for trouble report update: %v", herr)
		return herr
	}

	// Initialize dialog data for update operation
	dialogEditData := &EditDialogPageData{
		Submitted:      true, // Assume success initially
		ID:             int(id),
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	// Update trouble report if validation passes
	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		log.Infof("Updating trouble report %d by user %s: title=%q", id, user.UserName, title)

		// Retrieve original trouble report for change tracking
		trOld, err := db.TroubleReports.Get(id)
		if err != nil {
			log.Errorf("Failed to retrieve original trouble report %d: %v", id, err)
			return utils.HandlePgvisError(c, err)
		}

		// Create modification tracking metadata with original data
		modified := pgvis.NewModified[*pgvis.TroubleReport](user, trOld)
		trNew := pgvis.NewTroubleReport(modified, title, content)

		// Save updated report to database
		if err := db.TroubleReports.Update(id, trNew); err != nil {
			log.Errorf("Failed to update trouble report %d: %v", id, err)
			return utils.HandlePgvisError(c, err)
		}

		log.Infof("Successfully updated trouble report %d", id)
	} else {
		// Validation failed - keep dialog open for corrections
		dialogEditData.Submitted = false
		log.Debugf("Validation failed for trouble report %d update: title_invalid=%v, content_invalid=%v",
			id, dialogEditData.InvalidTitle, dialogEditData.InvalidContent)
	}

	// Return the dialog with updated state
	return GETDialogEdit(templates, c, db, dialogEditData)
}

// extractAndValidateFormData extracts title and content from form data,
// performs URL decoding, sanitization, and length validation.
//
// This function handles the common form processing logic shared between
// POST and PUT operations, ensuring consistent validation and error handling.
//
// Parameters:
//   - ctx: Echo context containing the form data
//
// Returns:
//   - title: Sanitized and validated title string
//   - content: Sanitized and validated content string
//   - httpErr: HTTP error if validation fails, nil on success
//
// Validation Rules:
//   - Title: 1-500 characters after sanitization
//   - Content: 1-50000 characters after sanitization
//   - Both fields are URL decoded and sanitized for security
func extractAndValidateFormData(ctx echo.Context) (title, content string, httpErr *echo.HTTPError) {
	var err error

	// Extract and decode title field
	title, err = url.QueryUnescape(ctx.FormValue(shared.TitleFormField))
	if err != nil {
		log.Warnf("Failed to decode title field: %v", err)
		return "", "", echo.NewHTTPError(http.StatusBadRequest,
			"Invalid title encoding: "+err.Error())
	}
	title = utils.SanitizeInput(title)

	// Extract and decode content field
	content, err = url.QueryUnescape(ctx.FormValue(shared.ContentFormField))
	if err != nil {
		log.Warnf("Failed to decode content field: %v", err)
		return "", "", echo.NewHTTPError(http.StatusBadRequest,
			"Invalid content encoding: "+err.Error())
	}
	content = utils.SanitizeInput(content)

	// Validate title length constraints
	if httpErr := utils.ValidateStringLength(title, shared.TitleFormField, shared.TitleMinLength, shared.TitleMaxLength); httpErr != nil {
		log.Debugf("Title validation failed: length=%d, min=%d, max=%d",
			len(title), shared.TitleMinLength, shared.TitleMaxLength)
		return title, content, httpErr
	}

	// Validate content length constraints
	if httpErr := utils.ValidateStringLength(content, shared.ContentFormField, shared.ContentMinLength, shared.ContentMaxLength); httpErr != nil {
		log.Debugf("Content validation failed: length=%d, min=%d, max=%d",
			len(content), shared.ContentMinLength, shared.ContentMaxLength)
		return title, content, httpErr
	}

	log.Debugf("Form validation successful: title_length=%d, content_length=%d",
		len(title), len(content))

	return title, content, nil
}
