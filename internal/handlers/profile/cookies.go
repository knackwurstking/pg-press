package profile

import (
	"slices"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/profile/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func HTMXGetCookies(c echo.Context) *echo.HTTPError {
	return renderCookies(c, false)
}

func HTMXDeleteCookies(c echo.Context) *echo.HTTPError {
	value, merr := utils.GetQueryString(c, "value")
	if merr != nil {
		return merr.Echo()
	}

	merr = db.DeleteCookie(value)
	if merr != nil {
		return merr.Echo()
	}

	eerr := HTMXGetCookies(c)
	if eerr != nil {
		return eerr
	}

	return renderCookies(c, true)
}

func renderCookies(c echo.Context, oob bool) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookies, merr := db.ListCookiesByApiKey(user.ApiKey)
	if merr != nil {
		return merr.Echo()
	}

	// Sort cookies by last login
	slices.SortFunc(cookies, func(a, b *shared.Cookie) int {
		return int(a.LastLogin - b.LastLogin)
	})

	t := templates.Cookies(templates.CookiesProps{
		Cookies: cookies,
		OOB:     oob,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cookies")
	}
	return nil
}
