package profile

import (
	"slices"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func HTMXGetCookies(c echo.Context) *echo.HTTPError {
	return renderCookies(c, false)
}

func HTMXDeleteCookies(c echo.Context) *echo.HTTPError {
	value, merr := shared.ParseQueryString(c, "value")
	if merr != nil {
		return merr.Echo()
	}

	merr = DB.User.Cookie.Delete(value)
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
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cookies, merr := helper.ListCookiesForApiKey(DB, user.ApiKey)
	if merr != nil {
		return merr.Echo()
	}

	// Sort cookies by last login
	slices.SortFunc(cookies, func(a, b *shared.Cookie) int {
		return int(a.LastLogin - b.LastLogin)
	})

	t := Cookies(CookiesProps{
		Cookies: cookies,
		OOB:     oob,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cookies")
	}
	return nil
}
