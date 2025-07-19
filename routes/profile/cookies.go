package profile

import (
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// GETCookies handles GET requests for the cookies page.
// It retrieves and displays all cookies associated with the authenticated user's API key.
func GETCookies(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	cookies, err := db.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	cookies = pgvis.SortCookies(cookies)

	return utils.HandleTemplate(c, cookies,
		templates,
		[]string{
			shared.ProfileCookiesTemplatePath,
		},
	)
}

// DELETECookies handles DELETE requests for removing cookies.
// It removes a specific cookie by value and then refreshes the cookies page.
func DELETECookies(templates fs.FS, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	value := utils.SanitizeInput(ctx.QueryParam("value"))
	if value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "cookie value parameter is required")
	}

	if err := db.Cookies.Remove(value); err != nil {
		return utils.HandlePgvisError(ctx, err)
	}

	return GETCookies(templates, ctx, db)
}
