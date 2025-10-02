package main

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
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
					log().Info("Admin %s: %s %s", user.Name, method, uri)
				}
			} else {
				// Log unauthorized modification attempts
				if method != "GET" && !keyAuthSkipper(c) && status >= 400 {
					log().Warn("Unauthorized %s to %s from %s (HTTP %d)", method, uri, remoteIP, status)
				}
			}

			// Log slow requests for performance monitoring
			if latency > 2*time.Second {
				log().Warn("Slow request: %s %s took %v from %s", method, uri, latency, remoteIP)
			}

			// Log unusually large request bodies
			if contentLength > 50*1024*1024 { // 50MB
				log().Info("Large upload: %s %s (%.1fMB) from %s", method, uri, float64(contentLength)/1024/1024, remoteIP)
			}

			// Log HTTP request with status-appropriate level
			if status >= 500 {
				// Server errors are already logged in error handler with more context
				log().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)
			} else if status >= 400 {
				// Only log client errors for non-routine cases
				if status != 404 || method != "GET" {
					log().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)
				}
			} else {
				// Success cases - use debug level for routine requests
				log().HTTPRequest(status, method, uri, remoteIP, latency, userInfo)
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
			clog("Middleware: Auth").Info("Authentication required for %s %s from %s",
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
	l := clog("Middleware: Auth Validator")
	remoteIP := ctx.RealIP()

	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		if user, err = db.Users.GetUserFromApiKey(auth); err != nil {
			l.Info("Authentication failed from %s", remoteIP)
			return false, echo.NewHTTPError(utils.GetHTTPStatusCode(err), "failed to validate user from API key: "+err.Error())
		}
		l.Debug("API key auth successful for user %s from %s", user.Name, remoteIP)
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *database.DB) (*models.User, error) {
	l := clog("Middleware: Cookie Validation")

	remoteIP := ctx.RealIP()
	httpCookie, err := ctx.Cookie(constants.CookieName)
	if err != nil {
		l.Debug("No cookie found for request from %s: %v", remoteIP, err)
		return nil, fmt.Errorf("failed to get cookie: %s", err.Error())
	}

	l.Debug("Found cookie for request from %s: expires=%v, secure=%v",
		remoteIP, httpCookie.Expires, httpCookie.Secure)

	cookie, err := db.Cookies.Get(httpCookie.Value)
	if err != nil {
		l.Debug("Cookie not found in database for request from %s: %v", remoteIP, err)
		return nil, fmt.Errorf("failed to get cookie: %s", err.Error())
	}

	l.Debug("Cookie found in database for request from %s: lastLogin=%d, userAgent='%s'",
		remoteIP, cookie.LastLogin, cookie.UserAgent)

	// Check if cookie has expired
	expirationTime := time.Now().Add(-constants.CookieExpirationDuration).UnixMilli()
	if cookie.LastLogin < expirationTime {
		l.Debug("Expired cookie from %s: lastLogin=%d, expirationTime=%d",
			remoteIP, cookie.LastLogin, expirationTime)
		return nil, utils.NewValidationError("cookie: cookie has expired")
	}

	l.Debug("Cookie is valid (not expired) for request from %s", remoteIP)

	user, err := db.Users.GetUserFromApiKey(cookie.ApiKey)
	if err != nil {
		l.Error("Failed to get user for cookie from %s: %v", remoteIP, err)
		return nil, fmt.Errorf("failed to validate user from API key: %s", err.Error())
	}

	l.Debug("User validated from cookie for request from %s: user=%s", remoteIP, user.Name)

	// Log user agent mismatch as potential security concern
	// Be more lenient for PWA compatibility - only log significant changes
	requestUserAgent := ctx.Request().UserAgent()
	if cookie.UserAgent != requestUserAgent {
		// Only log if the change seems significant (different browser/version, not just PWA vs browser mode)
		if !isMinorUserAgentChange(cookie.UserAgent, requestUserAgent) {
			l.Info("Significant user agent change for %s from %s", user.Name, remoteIP)
		} else {
			l.Debug("Minor user agent variation for %s from %s (likely PWA)", user.Name, remoteIP)
		}
	}

	if slices.Contains(pages, ctx.Request().URL.Path) {
		now := time.Now()
		cookie.LastLogin = now.UnixMilli()
		httpCookie.Expires = now.Add(constants.CookieExpirationDuration)

		l.Debug("Updating cookie for page visit by user %s from %s: path=%s",
			user.Name, remoteIP, ctx.Request().URL.Path)

		if err := db.Cookies.Update(cookie.Value, cookie); err != nil {
			l.Error("Failed to update cookie for user %s from %s: %v", user.Name, remoteIP, err)
		} else {
			l.Debug("Cookie successfully updated for user %s from %s", user.Name, remoteIP)
		}
	}

	l.Debug("Cookie validation successful for user %s from %s", user.Name, remoteIP)
	return user, nil
}

// isMinorUserAgentChange checks if the user agent change is minor (e.g., PWA vs browser mode)
// rather than a significant change (different browser/device)
func isMinorUserAgentChange(originalUA, newUA string) bool {
	if originalUA == newUA {
		return true
	}

	// If either is empty, consider it a significant change
	if originalUA == "" || newUA == "" {
		return false
	}

	// Common PWA-related user agent variations that should be considered minor:
	// - Addition or removal of "wv" (WebView)
	// - Changes in Chrome version numbers
	// - Addition/removal of PWA-specific identifiers

	// Extract the core browser identifier (Chrome, Firefox, Safari, etc.)
	originalCore := extractBrowserCore(originalUA)
	newCore := extractBrowserCore(newUA)

	// If the core browser is the same, consider it a minor change
	return originalCore != "" && originalCore == newCore
}

// extractBrowserCore extracts the core browser identifier from a user agent string
func extractBrowserCore(userAgent string) string {
	// Look for common browser patterns
	patterns := []string{
		"Chrome/", "Firefox/", "Safari/", "Edge/", "Opera/",
	}

	for _, pattern := range patterns {
		if pos := regexp.MustCompile(pattern).FindStringIndex(userAgent); pos != nil {
			// Return the browser name without version
			return pattern[:len(pattern)-1]
		}
	}

	return ""
}
