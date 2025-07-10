package html

import (
	"net/http"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

func handleLogout(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	if cookie, err := ctx.Cookie(CookieName); err == nil {
		db.Cookies.Remove(cookie.Value)

		err = ctx.Redirect(http.StatusSeeOther, "./login")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return nil
}
