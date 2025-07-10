package html

import (
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

func ServeLogout(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/logout", func(c echo.Context) error {
		return handleLogoutPage(c, options.DB)
	})
}

func handleLogoutPage(ctx echo.Context, db *pgvis.DB) error {
	if cookie, err := ctx.Cookie(CookieName); err == nil {
		db.Cookies.Remove(cookie.Value)

		err = ctx.Redirect(http.StatusSeeOther, "./login")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return nil
}
