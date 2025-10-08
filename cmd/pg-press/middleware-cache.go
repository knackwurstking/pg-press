package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// staticCacheMiddleware returns middleware that adds appropriate cache headers
// for static files (CSS, JS, images, etc.)
func staticCacheMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Remove server path prefix if present
			if serverPathPrefix != "" {
				path = strings.TrimPrefix(path, serverPathPrefix)
			}

			// Set cache headers based on file type
			if shouldCache(path) {
				setCacheHeaders(c, path)
			}

			return next(c)
		}
	}
}

// shouldCache determines if a file should be cached based on its extension
func shouldCache(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Cacheable file types
	cacheableExts := map[string]bool{
		".css":   true,
		".js":    true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".svg":   true,
		".ico":   true,
		".woff":  true,
		".woff2": true,
		".ttf":   true,
		".eot":   true,
		".json":  true, // for manifest.json
	}

	return cacheableExts[ext]
}

// setCacheHeaders sets appropriate cache headers based on file type
func setCacheHeaders(c echo.Context, path string) {
	ext := strings.ToLower(filepath.Ext(path))
	response := c.Response()

	// Different caching strategies for different file types
	switch ext {
	case ".css", ".js":
		// CSS and JS files - cache for 1 year since they should be versioned
		response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		response.Header().Set("Expires", time.Now().Add(365*24*time.Hour).Format(http.TimeFormat))

	case ".woff", ".woff2", ".ttf", ".eot":
		// Font files - cache for 1 year
		response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		response.Header().Set("Expires", time.Now().Add(365*24*time.Hour).Format(http.TimeFormat))

	case ".png", ".jpg", ".jpeg", ".gif", ".svg":
		// Images - cache for 30 days
		response.Header().Set("Cache-Control", "public, max-age=2592000")
		response.Header().Set("Expires", time.Now().Add(30*24*time.Hour).Format(http.TimeFormat))

	case ".ico":
		// Favicon - cache for 1 week
		response.Header().Set("Cache-Control", "public, max-age=604800")
		response.Header().Set("Expires", time.Now().Add(7*24*time.Hour).Format(http.TimeFormat))

	case ".json":
		// JSON files like manifest.json - cache for 1 day
		response.Header().Set("Cache-Control", "public, max-age=86400")
		response.Header().Set("Expires", time.Now().Add(24*time.Hour).Format(http.TimeFormat))

	default:
		// Other files - cache for 1 hour
		response.Header().Set("Cache-Control", "public, max-age=3600")
		response.Header().Set("Expires", time.Now().Add(time.Hour).Format(http.TimeFormat))
	}

	// Add ETag for better cache validation
	if etag := generateETag(path); etag != "" {
		response.Header().Set("ETag", etag)
	}

	// Add Vary header for proper cache behavior with compression
	response.Header().Set("Vary", "Accept-Encoding")

	// Set Last-Modified header (using current time as placeholder)
	// In a real implementation, you'd use the actual file modification time
	response.Header().Set("Last-Modified", time.Now().Add(-24*time.Hour).Format(http.TimeFormat))
}

// generateETag generates a simple ETag based on the file path
// In a production environment, you might want to use actual file content hash
func generateETag(path string) string {
	hasher := md5.New()
	hasher.Write([]byte(path))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	return fmt.Sprintf(`"pg-press-%s"`, hash[:8])
}

// assetVersion generates a version string for assets
// This can be used in templates to add version parameters to asset URLs
func assetVersion() string {
	// Using a simple timestamp-based version
	return fmt.Sprintf("%d", time.Now().Unix()/3600) // Changes every hour
}

// getAssetVersionParam returns a version parameter for use in URLs
func getAssetVersionParam() string {
	version := assetVersion()
	return "v=" + version
}

// conditionalCacheMiddleware handles conditional requests (If-None-Match, If-Modified-Since)
func conditionalCacheMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Handle conditional requests
			if handleConditionalRequest(c) {
				return c.NoContent(http.StatusNotModified)
			}

			return next(c)
		}
	}
}

// handleConditionalRequest checks if the request is conditional and if the resource hasn't changed
func handleConditionalRequest(c echo.Context) bool {
	path := c.Request().URL.Path

	if !shouldCache(path) {
		return false
	}

	// Check If-None-Match header (ETag)
	if inm := c.Request().Header.Get("If-None-Match"); inm != "" {
		etag := generateETag(path)
		if etag != "" && inm == etag {
			return true
		}
	}

	// Check If-Modified-Since header
	if ims := c.Request().Header.Get("If-Modified-Since"); ims != "" {
		if t, err := http.ParseTime(ims); err == nil {
			// For simplicity, assume files haven't been modified in the last 24 hours
			modTime := time.Now().Add(-24 * time.Hour)
			if !modTime.After(t) {
				return true
			}
		}
	}

	return false
}
