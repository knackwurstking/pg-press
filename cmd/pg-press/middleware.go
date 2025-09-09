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
			userAgent := req.UserAgent()
			contentLength := req.ContentLength

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
					// Log error details for server errors
					if status >= 500 {
						logger.Server().Error("Server error in request %s %s: %v", method, uri, err)
					}
				} else {
					status = dberror.GetHTTPStatusCode(err)
					logger.Server().Warn("Request error %s %s: %v", method, uri, err)
				}
			}

			// Get user info if available
			userInfo := ""
			if user, ok := c.Get("user").(*usermodels.User); ok {
				userInfo = " " + user.String()
				// Log security-relevant actions for admin users
				if user.IsAdmin() && (method == "POST" || method == "PUT" || method == "DELETE") {
					logger.Server().Info("Admin action: %s %s by %s", method, uri, user.Name)
				}
			} else {
				// Log unauthenticated requests to sensitive endpoints
				if method != "GET" && !keyAuthSkipper(c) {
					logger.Server().Warn("Unauthenticated %s request to %s from %s", method, uri, remoteIP)
				}
			}

			// Log slow requests for performance monitoring
			if latency > 1*time.Second {
				logger.Server().Warn("Slow request: %s %s took %v (user-agent: %s)", method, uri, latency, userAgent)
			}

			// Log large request bodies
			if contentLength > 10*1024*1024 { // 10MB
				logger.Server().Info("Large request body: %s %s (%d bytes) from %s", method, uri, contentLength, remoteIP)
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
	remoteIP := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		if user, err = db.Users.GetUserFromApiKey(auth); err != nil {
			logger.Middleware().Warn("Authentication failed from %s: invalid credentials (user-agent: %s)", remoteIP, userAgent)
			return false, echo.NewHTTPError(
				dberror.GetHTTPStatusCode(dberror.ErrInvalidCredentials),
				"failed to validate user from API key: "+err.Error())
		}
		logger.Middleware().Info("API key authentication successful for user %s from %s", user.Name, remoteIP)
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *database.DB) (*usermodels.User, error) {
	remoteIP := ctx.RealIP()

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
		logger.Middleware().Info("Expired cookie from %s (last login: %s)", remoteIP, c.TimeString())
		return nil, dberror.NewValidationError("cookie", "cookie has expired", nil)
	}

	user, err := db.Users.GetUserFromApiKey(c.ApiKey)
	if err != nil {
		logger.Middleware().Error("Failed to get user for cookie from %s: %v", remoteIP, err)
		return nil, fmt.Errorf("failed to validate user from API key: %s", err.Error())
	}

	// Log user agent mismatch as potential security concern
	requestUserAgent := ctx.Request().UserAgent()
	if c.UserAgent != requestUserAgent {
		logger.Middleware().Warn("User agent mismatch for user %s from %s: cookie=%s, request=%s",
			user.Name, remoteIP, c.UserAgent, requestUserAgent)
	}

	if slices.Contains(pages, ctx.Request().URL.Path) {

		now := time.Now()
		c.LastLogin = now.UnixMilli()
		cookie.Expires = now.Add(constants.CookieExpirationDuration)

		if err := db.Cookies.Update(c.Value, c); err != nil {
			logger.Middleware().Error("Failed to update cookie for user %s from %s: %v", user.Name, remoteIP, err)
		}
	}

	return user, nil
}
