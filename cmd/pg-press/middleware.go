package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/logger"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	logMiddleware      *ui.Logger
	keyAuthFilesToSkip []string
	pages              []string
)

func init() {
	logMiddleware = logger.New("middleware")

	// NOTE: Used for updating cookies
	pages = []string{
		env.ServerPathPrefix + "",
		env.ServerPathPrefix + "/profile",
		//env.ServerPathPrefix + "/feed",
		//env.ServerPathPrefix + "/editor",
		//env.ServerPathPrefix + "/help",
		//env.ServerPathPrefix + "/trouble-reports",
		//env.ServerPathPrefix + "/notes",
		//env.ServerPathPrefix + "/tools",
		//env.ServerPathPrefix + "/tool",
		//env.ServerPathPrefix + "/press",
		//env.ServerPathPrefix + "/umbau",
		//env.ServerPathPrefix + "/press-regenerations",
	}

	// NOTE: Important for skipping key authentication
	keyAuthFilesToSkip = []string{
		// Pages
		env.ServerPathPrefix + "/login",

		// CSS
		env.ServerPathPrefix + "/css/bootstrap-icons.min.css",
		env.ServerPathPrefix + "/css/ui.min.css",
		env.ServerPathPrefix + "/css/main-layout.css",

		// Libraries
		env.ServerPathPrefix + "/js/htmx-v2.0.7.min.js",
		env.ServerPathPrefix + "/js/main-layout.js",

		// Fonts
		env.ServerPathPrefix + "/bootstrap-icons.woff",
		env.ServerPathPrefix + "/bootstrap-icons.woff2",

		// Icons
		env.ServerPathPrefix + "/apple-touch-icon-180x180.png",
		env.ServerPathPrefix + "/favicon.ico",
		env.ServerPathPrefix + "/icon.png",
		env.ServerPathPrefix + "/manifest.json",
		env.ServerPathPrefix + "/pwa-192x192.png",
		env.ServerPathPrefix + "/pwa-512x512.png",
		env.ServerPathPrefix + "/pwa-64x64.png",
	}
}

func middlewareKeyAuth() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: keyAuthSkipper,
		KeyLookup: "header:" + echo.HeaderAuthorization +
			",query:access_token,cookie:" + auth.CookieName,
		AuthScheme: "Bearer",
		Validator: func(auth string, ctx echo.Context) (bool, error) {
			return keyAuthValidator(auth, ctx)
		},
		ErrorHandler: func(err error, c echo.Context) error {
			logMiddleware.Error(
				"KeyAuth error: %v [method=%s, path=%s, real_ip=%s]",
				err, c.Request().Method, c.Request().URL.Path, c.RealIP(),
			)
			merr := urlb.RedirectTo(c, urlb.UrlLogin("", nil).Page)
			if merr != nil {
				return merr.Err()
			}
			return nil
		},
	})
}

func keyAuthSkipper(ctx echo.Context) bool {
	url := ctx.Request().URL.String()
	path := ctx.Request().URL.Path

	return slices.Contains(keyAuthFilesToSkip, path) ||
		slices.Contains(keyAuthFilesToSkip, url)
}

func keyAuthValidator(auth string, ctx echo.Context) (bool, error) {
	realIP := ctx.RealIP()

	user, err := validateUserFromCookie(ctx)
	if err != nil {
		logMiddleware.Warn(
			"Validate user from cookie failed: %v [real_ip=%s]",
			err, realIP,
		)

		// Try to get user directly from the API key
		var merr *errors.MasterError
		user, merr = db.GetUserByApiKey(auth)
		if merr != nil {
			return false, merr
		}
	}

	logMiddleware.Debug(
		"API-Key auth successful for %s [real_ip=%s]",
		user.Name, realIP,
	)

	ctx.Set("user", user)
	ctx.Set("user-name", user.Name)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context) (*shared.User, error) {
	realIP := ctx.RealIP()
	httpCookie, err := ctx.Cookie(auth.CookieName)
	if err != nil {
		return nil, errors.Wrap(err, "get cookie")
	}

	cookie, merr := db.GetCookie(httpCookie.Value)
	if merr != nil {
		return nil, merr.Wrap("get cookie").Err()
	}

	// Check if cookie has expired
	if cookie.IsExpired() {
		return nil, fmt.Errorf("cookie has expired")
	}

	user, merr := db.GetUser(cookie.UserID)
	if merr != nil {
		return nil, merr.Wrap("validate user from API key").Err()
	}

	// Check if the path matches any of the tracked pages (ignoring prefix and query parameters)
	pathMatches := false
	requestPath := strings.TrimRight(ctx.Request().URL.Path, "/")
	if slices.Contains(pages, requestPath) {
		pathMatches = true
	}

	// Update last login time and cookie expiration if the path matches
	if pathMatches {
		cookie.LastLogin = shared.NewUnixMilli(time.Now())
		httpCookie.Expires = cookie.ExpiredAtTime()

		merr = db.UpdateCookie(cookie)
		if merr != nil {
			logMiddleware.Error(
				"Failed to update cookie: %v [user_name=%s, real_ip=%s]",
				merr, user.Name, realIP,
			)
		}
	}

	return user, nil
}
