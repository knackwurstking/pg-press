package profile

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// GETCookies handles GET requests for the cookies page.
// It retrieves and displays all cookies associated with the authenticated user's API key.
func GETCookies(templates fs.FS, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	// Get authenticated user from context
	user, herr := utils.GetUserFromContext(ctx)
	if herr != nil {
		return herr
	}

	// Retrieve cookies for the user's API key
	cookies, err := db.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return utils.HandlePgvisError(ctx, err)
	}

	// Sort cookies for consistent display order
	cookies = pgvis.SortCookies(cookies)

	// Parse the template
	t, err := template.ParseFS(templates, shared.ProfileCookiesTemplatePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Execute template with cookies data
	if err := t.Execute(ctx.Response(), cookies); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// DELETECookies handles DELETE requests for removing cookies.
// It removes a specific cookie by value and then refreshes the cookies page.
func DELETECookies(templates fs.FS, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	// Get and validate cookie value parameter
	value := utils.SanitizeInput(ctx.QueryParam("value"))
	if value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "cookie value parameter is required")
	}

	// Remove the cookie from database
	if err := db.Cookies.Remove(value); err != nil {
		return utils.HandlePgvisError(ctx, err)
	}

	// Refresh the cookies page by calling GET handler
	return GETCookies(templates, ctx, db)
}
