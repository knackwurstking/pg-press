package auth

import (
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func GetLogout(c echo.Context) *echo.HTTPError {
	Log.Debug("Logout attempt from IP: %#v", c.RealIP())

	if cookie, err := c.Cookie(CookieName); err == nil {
		merr := DB.User.Cookies.Delete(cookie.Value)
		if merr != nil {
			Log.Warn("Failed to delete cookie from database: %v", merr)
		}
	}

	if merr := urlb.RedirectTo(c, urlb.UrlLogin("", nil).Page); merr != nil {
		return merr.Echo()
	}
	return nil
}
