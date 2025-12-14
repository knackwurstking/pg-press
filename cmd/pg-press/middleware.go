package main

import (
	"fmt"
	"log"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

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

func middlewareKeyAuth(db *common.DB) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: keyAuthSkipper,
		KeyLookup: "header:" + echo.HeaderAuthorization +
			",query:access_token,cookie:" + auth.CookieName,
		AuthScheme: "Bearer",
		Validator: func(auth string, ctx echo.Context) (bool, error) {
			return keyAuthValidator(auth, ctx, db)
		},
		ErrorHandler: func(err error, c echo.Context) error {
			log.Println("KeyAuth error:", err, "Method:", c.Request().Method, "Path:", c.Request().URL.Path, "RealIP:", c.RealIP())
			merr := urlb.RedirectTo(c, urlb.UrlLogin("", nil).Page)
			if merr != nil {
				return merr.Err
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

func keyAuthValidator(auth string, ctx echo.Context, db *common.DB) (bool, error) {
	realIP := ctx.RealIP()

	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		log.Println("Validate user from cookie failed:", err, "RealIP:", realIP)
		// Try to get user directly from the API key
		users, merr := db.User.User.List()
		if merr != nil {
			return false, merr.Err
		}

		// Find user by API key
		var foundUser *shared.User
		for _, u := range users {
			if u.ApiKey == auth {
				foundUser = u
				break
			}
		}

		if foundUser == nil {
			return false, fmt.Errorf("invalid API key")
		}
		user = foundUser
	}

	if env.Verbose {
		log.Println("API key auth successful for user:", user.Name, "RealIP:", realIP)
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *common.DB) (*shared.User, error) {
	realIP := ctx.RealIP()
	httpCookie, err := ctx.Cookie(auth.CookieName)
	if err != nil {
		return nil, errors.Wrap(err, "get cookie")
	}

	cookie, merr := db.User.Cookie.GetByID(httpCookie.Value)
	if merr != nil {
		return nil, merr.Wrap("get cookie").Err
	}

	// Check if cookie has expired
	if cookie.IsExpired() {
		return nil, fmt.Errorf("cookie has expired")
	}

	user, merr := db.User.User.GetByID(cookie.UserID)
	if merr != nil {
		return nil, merr.Wrap("validate user from API key").Err
	}

	// Check if the path matches any of the tracked pages (ignoring prefix and query parameters)
	pathMatches := false
	requestPath := strings.TrimRight(ctx.Request().URL.Path, "/")
	if slices.Contains(pages, requestPath) {
		pathMatches = true
	}

	if pathMatches {
		cookie.LastLogin = shared.NewUnixMilli(time.Now())
		httpCookie.Expires = cookie.ExpiredAtTime()

		slog.Info("Updating cookie", "user_name", user.Name, "real_ip", realIP, "url_path", ctx.Request().URL.Path)

		// Try to update cookie with lock
		merr = db.User.Cookie.Update(cookie)
		if merr != nil {
			log.Println("Failed to update cookie with lock:", merr, "UserName:", user.Name, "RealIP:", realIP)
		}
	}

	return user, nil
}
