package auth

import (
	"log/slog"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func GetLogout(c echo.Context) *echo.HTTPError {
	slog.Debug("Logout attempt from IP",
		"real_ip", c.RealIP())

	if cookie, err := c.Cookie(CookieName); err == nil {
		merr := db.DeleteCookie(cookie.Value)
		if merr != nil {
			slog.Warn("Failed to delete cookie from database", "error", merr)
		}
	}

	if merr := utils.RedirectTo(c, urlb.Login("", nil)); merr != nil {
		return merr.Echo()
	}
	return nil
}
