// NOTE: Cleaned up by AI
package main

import (
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes"
)

var (
	keyAuthSkipperRegExp = regexp.MustCompile(
		`(.*/login.*|.*\.css|manifest.json|.*\.png|.*\.ico|.*service-worker\.js|.*\.woff|.*\.woff2|.*htmx.min.js|.*sw-register.js)`,
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
		Format: "${status} ${method} ${uri} (${remote_ip}) ${latency_human}\n",
		Output: os.Stderr,
	})
}

func middlewareKeyAuth(db *pgvis.DB) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper:    keyAuthSkipper,
		KeyLookup:  "header:" + echo.HeaderAuthorization + ",query:access_token,cookie:" + routes.CookieName,
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
		log.Debugf("Skipping authorization for url: %s", url)
		return true
	}
	return false
}

func keyAuthValidator(auth string, ctx echo.Context, db *pgvis.DB) (bool, error) {
	user, err := db.Users.GetUserFromApiKey(auth)
	if err != nil {
		user, err = validateUserFromCookie(ctx, db)
		if err != nil {
			return false, nil
		}
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *pgvis.DB) (*pgvis.User, error) {
	cookie, err := ctx.Cookie(routes.CookieName)
	if err != nil {
		return nil, err
	}

	c, err := db.Cookies.Get(cookie.Value)
	if err != nil {
		return nil, err
	}

	user, err := db.Users.GetUserFromApiKey(c.ApiKey)
	if err != nil {
		return nil, err
	}

	if slices.Contains(pages, ctx.Request().URL.Path) {
		log.Debugf("Updating cookies last login timestamp for user %s", user)

		c.LastLogin = time.Now().UnixMilli()
		cookie.Expires = time.Now().Add(routes.CookieExpirationDuration)

		if err := db.Cookies.Update(c.Value, c); err != nil {
			log.Errorf("Failed to update cookie for user %s: %s", user, err)
		}
	}

	return user, nil
}
