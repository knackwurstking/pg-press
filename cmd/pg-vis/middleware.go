package main

import (
	"bytes"
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
)

var (
	keyAuthSkipperRegExp = regexp.MustCompile(
		`(.*/login.*|.*\.css|.*\.png|.*\.ico|.*\.woff|.*\.woff2|.*manifest.json|.*service-worker\.js|.*htmx.min.js|.*sw-register.js|.*pwa-manager.js)`,
	)

	pages []string
)

func init() {
	pages = []string{
		serverPathPrefix + "/",
		serverPathPrefix + "/feed",
		serverPathPrefix + "/profile",
		serverPathPrefix + "/trouble-reports",
	}
}

func middlewareLogger() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${status} ${method} ${uri} (${remote_ip}) ${latency_human} ${custom}\n",
		Output: os.Stderr,
		CustomTagFunc: func(c echo.Context, buf *bytes.Buffer) (int, error) {
			if !slices.Contains(pages, c.Request().URL.Path) {
				return 0, nil
			}

			user, ok := c.Get("user").(*pgvis.User)
			if !ok {
				return 0, nil
			}

			return buf.WriteString(user.String())
		},
	})
}

func middlewareKeyAuth(db *pgvis.DB) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper:    keyAuthSkipper,
		KeyLookup:  "header:" + echo.HeaderAuthorization + ",query:access_token,cookie:" + constants.CookieName,
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
	if keyAuthSkipperRegExp.MatchString(url) {
		return true
	}
	return false
}

func keyAuthValidator(auth string, ctx echo.Context, db *pgvis.DB) (bool, error) {
	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		log.Warn("failed to validate user from cookie", "error", err)
		user, err = db.Users.GetUserFromApiKey(auth)
		if err != nil {
			return false, pgvis.WrapError(err, "failed to validate user from cookie")
		}
	}

	ctx.Set("user", user)
	return true, nil
}

// TODO: Checking the expiration time of the cookie
func validateUserFromCookie(ctx echo.Context, db *pgvis.DB) (*pgvis.User, error) {
	cookie, err := ctx.Cookie(constants.CookieName)
	if err != nil {
		return nil, pgvis.WrapError(err, "failed to get cookie")
	}

	c, err := db.Cookies.Get(cookie.Value)
	if err != nil {
		return nil, pgvis.WrapError(err, "failed to get cookie value")
	}

	user, err := db.Users.GetUserFromApiKey(c.ApiKey)
	if err != nil {
		return nil, pgvis.WrapError(err, "failed to get user from API key")
	}

	if slices.Contains(pages, ctx.Request().URL.Path) {
		log.Debugf("Updating cookies last login timestamp for user %s", user)

		c.LastLogin = time.Now().UnixMilli()
		cookie.Expires = time.Now().Add(constants.CookieExpirationDuration)

		if err := db.Cookies.Update(c.Value, c); err != nil {
			log.Errorf("Failed to update cookie for user %s: %s", user, err)
		}
	}

	return user, nil
}
