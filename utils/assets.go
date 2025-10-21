package utils

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/a-h/templ"
)

var (
	// assetVersionCache stores the computed asset version to avoid recomputation
	assetVersionCache string
	// assetVersionComputed tracks if we've already computed the version
	assetVersionComputed bool
)

// AssetVersion returns a version string for cache-busting assets
func AssetVersion() string {
	if assetVersionComputed {
		return assetVersionCache
	}

	// Generate version based on startup time
	// This ensures version changes on each deployment/restart
	startupTime := time.Now().Unix()
	assetVersionCache = fmt.Sprintf("%d", startupTime)
	assetVersionComputed = true
	return assetVersionCache
}

// AssetURL generates a versioned URL for an asset
func AssetURL(serverPathPrefix, assetPath string) templ.SafeURL {
	version := AssetVersion()
	baseURL := serverPathPrefix + assetPath

	if version != "" {
		return templ.SafeURL(fmt.Sprintf("%s?v=%s", baseURL, version))
	}

	return templ.URL(baseURL)
}

// AssetURLWithHash generates a URL with content-based hash (for stronger caching)
func AssetURLWithHash(serverPathPrefix, assetPath, content string) string {
	hasher := md5.New()
	hasher.Write([]byte(content))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

	baseURL := serverPathPrefix + assetPath
	return fmt.Sprintf("%s?h=%s", baseURL, hash)
}

// GetCacheHeaders returns appropriate cache control headers for different asset types
func GetCacheHeaders(assetPath string) map[string]string {
	headers := make(map[string]string)

	// Set appropriate cache headers based on file extension
	switch {
	case isStaticAsset(assetPath):
		// CSS, JS, fonts - cache for 1 year since they're versioned
		headers["Cache-Control"] = "public, max-age=31536000, immutable"
		headers["Expires"] = time.Now().Add(365 * 24 * time.Hour).Format(time.RFC1123)

	case isImage(assetPath):
		// Images - cache for 30 days
		headers["Cache-Control"] = "public, max-age=2592000"
		headers["Expires"] = time.Now().Add(30 * 24 * time.Hour).Format(time.RFC1123)

	case isIcon(assetPath):
		// Icons and favicons - cache for 1 week
		headers["Cache-Control"] = "public, max-age=604800"
		headers["Expires"] = time.Now().Add(7 * 24 * time.Hour).Format(time.RFC1123)

	default:
		// Default - cache for 1 hour
		headers["Cache-Control"] = "public, max-age=3600"
		headers["Expires"] = time.Now().Add(time.Hour).Format(time.RFC1123)
	}

	headers["Vary"] = "Accept-Encoding"
	return headers
}

// isStaticAsset checks if the asset is CSS, JS, or font file
func isStaticAsset(path string) bool {
	extensions := []string{".css", ".js", ".woff", ".woff2", ".ttf", ".eot"}
	for _, ext := range extensions {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isImage checks if the asset is an image file
func isImage(path string) bool {
	extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp"}
	for _, ext := range extensions {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isIcon checks if the asset is an icon file
func isIcon(path string) bool {
	extensions := []string{".ico"}
	iconNames := []string{"favicon", "apple-touch-icon", "maskable-icon", "pwa-"}

	// Check extensions
	for _, ext := range extensions {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}

	// Check common icon file names
	for _, name := range iconNames {
		if contains(path, name) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr))))
}

// containsInMiddle checks if substring exists in the middle of string
func containsInMiddle(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
