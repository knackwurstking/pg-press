// TODO: Move middleware to internal/middleware
package main

import (
	"bytes"
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/knackwurstking/pgpress/internal/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
)

var (
	// FIXME: Do not use regexp for this
	keyAuthSkipperRegExp = regexp.MustCompile(
		`(.*/login.*|.*\.css|.*\.png|.*\.ico|.*\.woff|.*\.woff2|.*manifest.json|` +
			`.*service-worker\.js|.*htmx.min.js|.*sw-register.js|.*pwa-manager.js)`)

	pages = []string{
		serverPathPrefix + "/",
		serverPathPrefix + "/feed",
		serverPathPrefix + "/profile",
		serverPathPrefix + "/trouble-reports",
	}
)

func middlewareLogger() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${status} ${method} ${uri} (${remote_ip}) " +
			"${latency_human} ${custom}\n",
		Output: os.Stderr,
		CustomTagFunc: func(c echo.Context, buf *bytes.Buffer) (int, error) {
			if !slices.Contains(pages, c.Request().URL.Path) {
				return 0, nil
			}

			user, ok := c.Get("user").(*database.User)
			if !ok {
				return 0, nil
			}

			return buf.WriteString(user.String())
		},
	})
}

func middlewareKeyAuth(db *database.DB) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: keyAuthSkipper,
		KeyLookup: "header:" + echo.HeaderAuthorization +
			",query:access_token,cookie:" + constants.CookieName,
		AuthScheme: "Bearer",
		Validator: func(auth string, ctx echo.Context) (bool, error) {
			return keyAuthValidator(auth, ctx, db)
		},
		ErrorHandler: func(err error, c echo.Context) error {
			return c.Redirect(http.StatusSeeOther, serverPathPrefix+"/login")
		},
	})
}

func keyAuthSkipper(ctx echo.Context) bool {
	url := ctx.Request().URL.String()
	return keyAuthSkipperRegExp.MatchString(url)
}

func keyAuthValidator(auth string, ctx echo.Context, db *database.DB) (bool, error) {
	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		logger.Middleware().Warn("failed to validate user from cookie: %v", err)
		if user, err = db.Users.GetUserFromApiKey(auth); err != nil {
			return false, err
		}
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *database.DB) (*database.User, error) {
	cookie, err := ctx.Cookie(constants.CookieName)
	if err != nil {
		return nil, err
	}

	c, err := db.Cookies.Get(cookie.Value)
	if err != nil {
		return nil, err
	}

	// Check if cookie has expired
	expirationTime := time.Now().Add(-constants.CookieExpirationDuration).UnixMilli()
	if c.LastLogin < expirationTime {
		return nil, database.NewValidationError("cookie", "cookie has expired", nil)
	}

	user, err := db.Users.GetUserFromApiKey(c.ApiKey)
	if err != nil {
		return nil, err
	}

	if slices.Contains(pages, ctx.Request().URL.Path) {
		logger.Middleware().Info(
			"Updating cookies last login timestamp for user %s", user)
		now := time.Now()
		c.LastLogin = now.UnixMilli()
		cookie.Expires = now.Add(constants.CookieExpirationDuration)

		if err := db.Cookies.Update(c.Value, c); err != nil {
			logger.Middleware().Error(
				"Failed to update cookie for user %s: %v", user, err)
		}
	}

	return user, nil
}
