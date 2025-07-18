// Package profile provides HTTP route handlers for user profile management.
//
// This package implements the web interface for viewing and managing user profiles,
// including personal information, authentication cookies, and account settings.
// It handles both the main profile page rendering and HTMX-based interactions
// for dynamic updates without full page reloads.
//
// Key Features:
//   - User profile display with editable information
//   - Username modification functionality
//   - Authentication cookie management and viewing
//   - Session cleanup and deletion capabilities
//   - Integration with the pgvis database layer
package profile

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

// Profile contains the data structure passed to the profile page template.
// This structure includes user information and associated session data needed
// for rendering the profile interface and management controls.
type Profile struct {
	// User contains the authenticated user's information
	User *pgvis.User `json:"user"`

	// Cookies contains all authentication sessions/cookies for the user
	Cookies []*pgvis.Cookie `json:"cookies"`
}

// CookiesSorted returns the user's cookies sorted by last login time.
// This method provides a consistent ordering for display purposes,
// with the most recently used sessions appearing first.
//
// Returns:
//   - []*pgvis.Cookie: Sorted slice of cookies by last login (newest first)
func (p *Profile) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

// Serve configures and registers all profile related HTTP routes.
//
// This function sets up the following endpoints:
//   - GET    /profile         - Main profile page with user information
//   - GET    /profile/cookies - Cookie/session data listing (HTMX)
//   - DELETE /profile/cookies - Delete specific cookie/session (HTMX)
//
// Parameters:
//   - templates: Embedded filesystem containing HTML templates
//   - serverPathPrefix: URL prefix for all routes (e.g., "/app")
//   - e: Echo instance to register routes with
//   - db: Database connection for data operations
func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	// Main profile page
	e.GET(serverPathPrefix+"/profile", handleMainPage(templates, db))

	// Cookie management endpoints (HTMX)
	cookiesPath := serverPathPrefix + "/profile/cookies"
	e.GET(cookiesPath, handleGetCookies(templates, db))
	e.DELETE(cookiesPath, handleDeleteCookies(templates, db))
}

// handleMainPage returns a handler for the main profile page.
// This page displays user information, allows username editing, and shows
// active authentication sessions with management capabilities.
func handleMainPage(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Initialize profile data structure
		pageData := &Profile{
			Cookies: make([]*pgvis.Cookie, 0),
		}

		// Authenticate user and retrieve context information
		user, herr := utils.GetUserFromContext(c)
		if herr != nil {
			log.Warnf("Failed to get user from context for profile page: %v", herr)
			return herr
		}
		pageData.User = user

		// Handle username change if submitted
		if err := handleUserNameChange(c, pageData, db); err != nil {
			log.Errorf("Failed to handle username change: %v", err)
			return utils.HandlePgvisError(c, err)
		}

		// Load user's authentication cookies/sessions
		if cookies, err := db.Cookies.ListApiKey(pageData.User.ApiKey); err != nil {
			log.Errorf("Failed to retrieve cookies for user %s: %v", user.UserName, err)
			// Continue with empty cookies list rather than failing completely
		} else {
			pageData.Cookies = cookies
			log.Debugf("Loaded %d cookies for user %s", len(cookies), user.UserName)
		}

		// Parse required templates for the profile page
		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.ProfileTemplatePath,
			shared.NavFeedTemplatePath,
		)
		if err != nil {
			log.Errorf("Failed to parse profile page templates: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to load page templates: "+err.Error())
		}

		// Render the profile page with user data
		if err = t.Execute(c.Response(), pageData); err != nil {
			log.Errorf("Failed to execute profile page template: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError,
				"Failed to render page: "+err.Error())
		}

		return nil
	}
}

// handleGetCookies returns a handler for the cookies/sessions data endpoint.
// This endpoint is used by HTMX to load and refresh the cookies table
// without requiring a full page reload.
func handleGetCookies(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return GETCookies(templates, c, db)
	}
}

// handleDeleteCookies returns a handler for deleting authentication cookies/sessions.
// This endpoint allows users to terminate specific sessions or clean up
// expired authentication tokens.
func handleDeleteCookies(templates fs.FS, db *pgvis.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return DELETECookies(templates, c, db)
	}
}

// handleUserNameChange processes username modification requests from the profile form.
// This function validates and applies username changes, ensuring data integrity
// and proper validation of the new username.
//
// The function performs the following operations:
//  1. Extracts the new username from form data
//  2. Validates the username against length constraints
//  3. Checks if the username has actually changed
//  4. Updates the user record in the database
//  5. Updates the profile data structure for immediate display
//
// Parameters:
//   - ctx: Echo context containing the form data
//   - pageData: Profile data structure to update with new information
//   - db: Database connection for user operations
//
// Returns:
//   - error: Error if validation or database operation fails, nil on success
//
// Validation Rules:
//   - Username must be between 1 and 100 characters
//   - Username is sanitized for security purposes
//   - Only updates if the username has actually changed
func handleUserNameChange(ctx echo.Context, pageData *Profile, db *pgvis.DB) error {
	// Extract form parameters
	formParams, _ := ctx.FormParams()
	userName := utils.SanitizeInput(formParams.Get(shared.UserNameFormField))

	// Check if username change was requested and is different
	if userName == "" || userName == pageData.User.UserName {
		return nil // No change requested or same username
	}

	log.Debugf("Processing username change for user %d: %s => %s",
		pageData.User.TelegramID, pageData.User.UserName, userName)

	// Validate new username length
	if len(userName) < shared.UserNameMinLength || len(userName) > shared.UserNameMaxLength {
		log.Warnf("Invalid username length: %d characters (min: %d, max: %d)",
			len(userName), shared.UserNameMinLength, shared.UserNameMaxLength)
		return pgvis.NewValidationError(shared.UserNameFormField,
			"username must be between 1 and 100 characters", len(userName))
	}

	// Create updated user record
	updatedUser := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)

	// Update user in database
	if err := db.Users.Update(pageData.User.TelegramID, updatedUser); err != nil {
		log.Errorf("Failed to update username in database: %v", err)
		return err
	}

	// Update the profile data for immediate display
	pageData.User.UserName = userName

	log.Infof("Successfully updated username for user %d to: %s",
		pageData.User.TelegramID, userName)

	return nil
}
