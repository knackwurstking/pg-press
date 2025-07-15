package profile

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/html/handler"
	"github.com/knackwurstking/pg-vis/pgvis"
)

func GETCookies(templates fs.FS, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := handler.GetUserFromContext(ctx)
	if herr != nil {
		return herr
	}

	cookies, err := db.Cookies.ListApiKey(user.ApiKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("list cookies for api key failed: %s", err))
	}
	cookies = pgvis.SortCookies(cookies)

	t, err := template.ParseFS(templates, "templates/profile/cookies.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("template parsing failed: %s", err))
	}

	if err := t.Execute(ctx.Response(), cookies); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("template executing failed: %s", err))
	}

	return nil
}

func DELETECookies(templates fs.FS, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	value := ctx.QueryParam("value")
	if value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("query \"value\" is missing"))
	}

	if err := db.Cookies.Remove(value); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("removing cookie failed: %s", err))
	}

	return GETCookies(templates, ctx, db)
}
