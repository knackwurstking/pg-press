package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	keyAuthFilesToSkip []string
	pages              []string
)

func init() {
	// NOTE: Used for updating cookies
	pages = []string{
		env.ServerPathPrefix + "",
		env.ServerPathPrefix + "/profile",
		env.ServerPathPrefix + "/feed",
		env.ServerPathPrefix + "/editor",
		env.ServerPathPrefix + "/trouble-reports",
		env.ServerPathPrefix + "/notes",
		env.ServerPathPrefix + "/tools",
		env.ServerPathPrefix + "/tool",
		env.ServerPathPrefix + "/press",
		env.ServerPathPrefix + "/umbau",
	}

	// NOTE: Important for skipping key authentication
	keyAuthFilesToSkip = []string{
		// Pages
		env.ServerPathPrefix + "/login",

		// CSS
		env.ServerPathPrefix + "/css/output.css",

		// JS
		env.ServerPathPrefix + "/js/layout/htmx-v2.0.7.min.js",
		env.ServerPathPrefix + "/js/layout/main.js",

		// TemplUI Component Scripts
		env.ServerPathPrefix + "/js/components/checkbox.min.js",
		env.ServerPathPrefix + "/js/components/dialog.min.js",
		env.ServerPathPrefix + "/js/components/image-viewer.js",
		env.ServerPathPrefix + "/js/components/input.min.js",
		env.ServerPathPrefix + "/js/components/label.min.js",
		env.ServerPathPrefix + "/js/components/popover.min.js",
		env.ServerPathPrefix + "/js/components/selectbox.min.js",
		env.ServerPathPrefix + "/js/components/tabs.min.js",
		env.ServerPathPrefix + "/js/components/textarea.min.js",

		// Icons
		env.ServerPathPrefix + "/apple-touch-icon-180x180.png",
		env.ServerPathPrefix + "/favicon.ico",
		env.ServerPathPrefix + "/icon.png",
		env.ServerPathPrefix + "/manifest.json",
		env.ServerPathPrefix + "/maskable-icon-512x512.png",
		env.ServerPathPrefix + "/pwa-64x64.png",
		env.ServerPathPrefix + "/pwa-192x192.png",
		env.ServerPathPrefix + "/pwa-512x512.png",
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
			slog.Error("KeyAuth error",
				"error", err,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"real_ip", c.RealIP())

			merr := utils.RedirectTo(c, urlb.Login("", nil))
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
		slog.Warn("Validate user from cookie failed",
			"error", err,
			"real_ip", realIP)

		// Try to get user directly from the API key
		var merr *errors.HTTPError
		user, merr = db.GetUserByApiKey(auth)
		if merr != nil {
			return false, merr
		}
	}

	slog.Debug("API-Key auth successful",
		"user_name", user.Name,
		"real_ip", realIP)

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
			slog.Error("Failed to update cookie",
				"error", merr,
				"user_name", user.Name,
				"real_ip", realIP)
		}
	}

	return user, nil
}
