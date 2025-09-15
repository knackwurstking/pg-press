package main

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

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
			contentLength := req.ContentLength

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = utils.GetHTTPStatusCode(err)
				}
				// Error details are logged in the HTTP error handler
			}

			// Get user info if available
			userInfo := ""
			if user, ok := c.Get("user").(*models.User); ok {
				userInfo = " " + user.String()
				// Log critical admin actions
				if user.IsAdmin() && (method == "DELETE" || (method == "POST" && status < 400)) {
					logger.Server().Info("Admin %s: %s %s", user.Name, method, uri)
				}
			} else {
				// Log unauthorized modification attempts
				if method != "GET" && !keyAuthSkipper(c) && status >= 400 {
					logger.Server().Warn("Unauthorized %s to %s from %s (HTTP %d)", method, uri, remoteIP, status)
				}
			}

			// Log slow requests for performance monitoring
			if latency > 2*time.Second {
				logger.Server().Warn("Slow request: %s %s took %v from %s", method, uri, latency, remoteIP)
			}

			// Log unusually large request bodies
			if contentLength > 50*1024*1024 { // 50MB
				logger.Server().Info("Large upload: %s %s (%.1fMB) from %s", method, uri, float64(contentLength)/1024/1024, remoteIP)
			}

			// Log HTTP request with status-appropriate level
			if status >= 500 {
				// Server errors are already logged in error handler with more context
				logger.Server().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)
			} else if status >= 400 {
				// Only log client errors for non-routine cases
				if status != 404 || method != "GET" {
					logger.Server().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)
				}
			} else {
				// Success cases - use debug level for routine requests
				if latency > 500*time.Millisecond || method != "GET" {
					logger.Server().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)
				}
			}

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
			remoteIP := c.RealIP()
			logger.Middleware().Info("Authentication required for %s %s from %s",
				c.Request().Method, c.Request().URL.Path, remoteIP)
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

	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		if user, err = db.Users.GetUserFromApiKey(auth); err != nil {
			logger.Middleware().Info("Authentication failed from %s", remoteIP)
			return false, echo.NewHTTPError(utils.GetHTTPStatusCode(err), "failed to validate user from API key: "+err.Error())
		}
		logger.Middleware().Debug("API key auth successful for user %s from %s", user.Name, remoteIP)
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *database.DB) (*models.User, error) {
	remoteIP := ctx.RealIP()
	httpCookie, err := ctx.Cookie(constants.CookieName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookie: %s", err.Error())
	}

	cookie, err := db.Cookies.Get(httpCookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookie: %s", err.Error())
	}

	// Check if cookie has expired
	expirationTime := time.Now().Add(-constants.CookieExpirationDuration).UnixMilli()
	if cookie.LastLogin < expirationTime {
		logger.Middleware().Debug("Expired cookie from %s", remoteIP)
		return nil, utils.NewValidationError("cookie: cookie has expired")
	}

	user, err := db.Users.GetUserFromApiKey(cookie.ApiKey)
	if err != nil {
		logger.Middleware().Error("Failed to get user for cookie from %s: %v", remoteIP, err)
		return nil, fmt.Errorf("failed to validate user from API key: %s", err.Error())
	}

	// Log user agent mismatch as potential security concern
	requestUserAgent := ctx.Request().UserAgent()
	if cookie.UserAgent != requestUserAgent {
		logger.Middleware().Info("User agent changed for %s from %s", user.Name, remoteIP)
	}

	if slices.Contains(pages, ctx.Request().URL.Path) {
		now := time.Now()
		cookie.LastLogin = now.UnixMilli()
		httpCookie.Expires = now.Add(constants.CookieExpirationDuration)

		if err := db.Cookies.Update(cookie.Value, cookie); err != nil {
			logger.Middleware().Error("Failed to update cookie for user %s from %s: %v", user.Name, remoteIP, err)
		}
	}

	return user, nil
}
