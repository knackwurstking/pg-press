package profile

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

func GETCookies(templates fs.FS, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(ctx)
	if herr != nil {
		return herr
	}

	cookies, err := db.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return utils.HandlePgvisError(ctx, err)
	}
	cookies = pgvis.SortCookies(cookies)

	t, err := template.ParseFS(templates, "templates/profile/cookies.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := t.Execute(ctx.Response(), cookies); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

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
