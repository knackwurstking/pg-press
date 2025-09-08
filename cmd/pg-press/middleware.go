package main

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	keyAuthFilesToSkip       []string
	keyAuthFilesToSkipRegExp *regexp.Regexp
	pages                    []string
)

func init() {
	pages = []string{
		serverPathPrefix + "/",
		serverPathPrefix + "/feed",
		serverPathPrefix + "/profile",
		serverPathPrefix + "/trouble-reports",
	}

	keyAuthFilesToSkip = []string{
		// Pages
		serverPathPrefix + "/login",

		// CSS
		serverPathPrefix + "/css/bootstrap-icons.min.css",
		serverPathPrefix + "/css/ui.min.css",
		serverPathPrefix + "/css/layout.css",

		// Libraries
		serverPathPrefix + "/js/htmx-v2.0.6.min.js",
		serverPathPrefix + "/js/htmx-ext-ws-v2.0.3.min.js",

		// Fonts
		serverPathPrefix + "/bootstrap-icons.woff",
		serverPathPrefix + "/bootstrap-icons.woff2",

		// Icons
		serverPathPrefix + "/apple-touch-icon-180x180.png",
		serverPathPrefix + "/favicon.ico",
		serverPathPrefix + "/icon.png",
		serverPathPrefix + "/manifest.json",
		serverPathPrefix + "/pwa-192x192.png",
		serverPathPrefix + "/pwa-512x512.png",
		serverPathPrefix + "/pwa-64x64.png",
	}

	keyAuthFilesToSkipRegExp = regexp.MustCompile(`.*woff[2]`)
}

func middlewareLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Get request info
			req := c.Request()
			res := c.Response()
			status := res.Status
			method := req.Method
			uri := req.RequestURI
			remoteIP := c.RealIP()

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = dberror.GetHTTPStatusCode(err)
				}
			}

			// Get user info if available
			userInfo := ""
			if user, ok := c.Get("user").(*usermodels.User); ok {
				userInfo = " " + user.String()
			}

			// Log using internal logger with special HTTP highlighting
			logger.Server().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)

			return err
		}
	}
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
			logger.Middleware().Warn("Handling error, redirect user to the login page: %#v", err)
			return c.Redirect(http.StatusSeeOther, serverPathPrefix+"/login")
		},
	})
}

func keyAuthSkipper(ctx echo.Context) bool {
	url := ctx.Request().URL.String()
	path := ctx.Request().URL.Path
	if slices.Contains(keyAuthFilesToSkip, path) || slices.Contains(keyAuthFilesToSkip, url) {
		return true
	}
	return keyAuthFilesToSkipRegExp.MatchString(url)
}

func keyAuthValidator(auth string, ctx echo.Context, db *database.DB) (bool, error) {
	user, err := validateUserFromCookie(ctx, db)
	logger.Middleware().Debug("Received login request, apiKey: %#v, user: %#v, err: %v", auth, user, err)
	if err != nil {
		logger.Middleware().Warn("failed to validate user from cookie: %v", err)
		if user, err = db.Users.GetUserFromApiKey(auth); err != nil {
			return false, echo.NewHTTPError(
				dberror.GetHTTPStatusCode(dberror.ErrInvalidCredentials),
				"failed to validate user from API key: "+err.Error())
		}
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *database.DB) (*usermodels.User, error) {
	cookie, err := ctx.Cookie(constants.CookieName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookie: %s", err.Error())
	}

	c, err := db.Cookies.Get(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookie: %s", err.Error())
	}

	// Check if cookie has expired
	expirationTime := time.Now().Add(-constants.CookieExpirationDuration).UnixMilli()
	if c.LastLogin < expirationTime {
		return nil, dberror.NewValidationError("cookie", "cookie has expired", nil)
	}

	user, err := db.Users.GetUserFromApiKey(c.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user from API key: %s", err.Error())
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
