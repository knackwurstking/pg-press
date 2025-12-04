package main

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	keyAuthFilesToSkip       []string
	keyAuthFilesToSkipRegExp *regexp.Regexp
	pages                    []string
)

func init() {
	// NOTE: Used for updating cookies
	pages = []string{
		env.ServerPathPrefix + "/",
		env.ServerPathPrefix + "/feed",
		env.ServerPathPrefix + "/profile",
		env.ServerPathPrefix + "/editor",
		env.ServerPathPrefix + "/help",
		env.ServerPathPrefix + "/trouble-reports",
		env.ServerPathPrefix + "/notes",
		env.ServerPathPrefix + "/tools",
		env.ServerPathPrefix + "/tool",
		env.ServerPathPrefix + "/press",
		env.ServerPathPrefix + "/umbau",
		env.ServerPathPrefix + "/press-regenerations",
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
		env.ServerPathPrefix + "/js/htmx-ext-ws-v2.0.3.min.js",
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

	keyAuthFilesToSkipRegExp = regexp.MustCompile(`.*woff[2]`)
}

func middlewareKeyAuth(db *services.Registry) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: keyAuthSkipper,
		KeyLookup: "header:" + echo.HeaderAuthorization +
			",query:access_token,cookie:" + env.CookieName,
		AuthScheme: "Bearer",
		Validator: func(auth string, ctx echo.Context) (bool, error) {
			return keyAuthValidator(auth, ctx, db)
		},
		ErrorHandler: func(err error, c echo.Context) error {
			slog.Error(
				"Authentication required",
				"method", c.Request().Method,
				"url_path", c.Request().URL.Path,
				"real_ip", c.RealIP(),
			)
			return utils.RedirectTo(c, utils.UrlLogin("", nil).Page)
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

func keyAuthValidator(auth string, ctx echo.Context, db *services.Registry) (bool, error) {
	realIP := ctx.RealIP()

	user, err := validateUserFromCookie(ctx, db)
	if err != nil {
		if user, err = db.Users.GetUserFromApiKey(auth); err != nil {
			slog.Error("Authentication failed", "real_ip", realIP)
			return false, echo.NewHTTPError(errors.GetHTTPStatusCodeFromError(err), "validate user from API key: "+err.Error())
		}
		slog.Info("API key auth successful", "user_name", user.Name, "real_ip", realIP)
	}

	ctx.Set("user", user)
	return true, nil
}

func validateUserFromCookie(ctx echo.Context, db *services.Registry) (*models.User, error) {
	realIP := ctx.RealIP()
	httpCookie, err := ctx.Cookie(env.CookieName)
	if err != nil {
		return nil, errors.Wrap(err, "get cookie")
	}

	cookie, dberr := db.Cookies.Get(httpCookie.Value)
	if dberr != nil {
		return nil, errors.Wrap(dberr, "get cookie")
	}

	// Check if cookie has expired
	if cookie.IsExpired() {
		slog.Error("Cookie has expired", "real_ip", realIP)
		return nil, fmt.Errorf("cookie has expired")
	}

	user, err := db.Users.GetUserFromApiKey(cookie.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("validate user from API key: %s", err.Error())
	}

	// Log user agent mismatch as potential security concern
	// Be more lenient for PWA compatibility - only log significant changes
	requestUserAgent := ctx.Request().UserAgent()
	if cookie.UserAgent != requestUserAgent {
		// Only log if the change seems significant (different browser/version, not just PWA vs browser mode)
		if !isMinorUserAgentChange(cookie.UserAgent, requestUserAgent) {
			slog.Info("Significant user agent change", "user_name", user.Name, "real_ip", realIP)
		} else {
			slog.Info("Minor user agent variation (likely PWA)", "user_name", user.Name, "real_ip", realIP)
		}
	}

	// Check if the path matches any of the tracked pages (ignoring prefix and query parameters)
	pathMatches := false
	requestPath := ctx.Request().URL.Path
	for _, page := range pages {
		if strings.HasSuffix(requestPath, page) {
			pathMatches = true
			break
		}
	}

	if pathMatches {
		cookie.UpdateLastLogin()
		httpCookie.Expires = cookie.Expires()

		slog.Info("Updating cookie", "user_name", user.Name, "real_ip", realIP, "url_path", ctx.Request().URL.Path)

		// Try to update cookie with lock
		if err := db.Cookies.Update(cookie.Value, cookie); err != nil {
			// If the update failed due to a race condition (another goroutine updated first),
			// we log it but don't treat it as a critical error
			if strings.Contains(err.Error(), "cookie was updated by another process") {
				slog.Debug("Cookie update skipped due to race condition", "user_name", user.Name, "real_ip", realIP)
			} else {
				slog.Error("Failed to update cookie with lock", "user_name", user.Name, "real_ip", realIP, "error", err)
			}
		} else {
			slog.Debug("Cookie successfully updated with lock", "user_name", user.Name, "real_ip", realIP)
		}
	}

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
