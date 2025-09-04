package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestStaticCacheMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedCache  bool
		expectedMaxAge string
		expectedCC     string
	}{
		{
			name:           "CSS file should be cached for 1 year",
			path:           "/css/ui.min.css",
			expectedCache:  true,
			expectedMaxAge: "31536000",
			expectedCC:     "public, max-age=31536000, immutable",
		},
		{
			name:           "JS file should be cached for 1 year",
			path:           "/js/htmx-v2.0.6.min.js",
			expectedCache:  true,
			expectedMaxAge: "31536000",
			expectedCC:     "public, max-age=31536000, immutable",
		},
		{
			name:           "PNG image should be cached for 30 days",
			path:           "/images/icon.png",
			expectedCache:  true,
			expectedMaxAge: "2592000",
			expectedCC:     "public, max-age=2592000",
		},
		{
			name:           "Favicon should be cached for 1 week",
			path:           "/favicon.ico",
			expectedCache:  true,
			expectedMaxAge: "604800",
			expectedCC:     "public, max-age=604800",
		},
		{
			name:           "Font file should be cached for 1 year",
			path:           "/fonts/bootstrap-icons.woff2",
			expectedCache:  true,
			expectedMaxAge: "31536000",
			expectedCC:     "public, max-age=31536000, immutable",
		},
		{
			name:           "JSON file should be cached for 1 day",
			path:           "/manifest.json",
			expectedCache:  true,
			expectedMaxAge: "86400",
			expectedCC:     "public, max-age=86400",
		},
		{
			name:          "HTML file should not be cached aggressively",
			path:          "/index.html",
			expectedCache: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create a simple handler that just returns OK
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "test content")
			}

			// Apply middleware
			middleware := staticCacheMiddleware()
			err := middleware(handler)(c)

			// Assertions
			assert.NoError(t, err)

			if tt.expectedCache {
				// Check Cache-Control header
				cacheControl := rec.Header().Get("Cache-Control")
				assert.Equal(t, tt.expectedCC, cacheControl, "Cache-Control header should match")

				// Check Expires header exists and is in the future
				expires := rec.Header().Get("Expires")
				assert.NotEmpty(t, expires, "Expires header should be set")

				expiresTime, err := http.ParseTime(expires)
				assert.NoError(t, err, "Expires header should be valid")
				assert.True(t, expiresTime.After(time.Now()), "Expires should be in the future")

				// Check ETag header
				etag := rec.Header().Get("ETag")
				assert.NotEmpty(t, etag, "ETag header should be set")
				assert.True(t, strings.HasPrefix(etag, `"pg-vis-`), "ETag should have correct format")

				// Check Vary header
				vary := rec.Header().Get("Vary")
				assert.Equal(t, "Accept-Encoding", vary, "Vary header should be set")

				// Check Last-Modified header
				lastModified := rec.Header().Get("Last-Modified")
				assert.NotEmpty(t, lastModified, "Last-Modified header should be set")
			} else {
				// For non-cacheable files, headers should not be aggressively set
				cacheControl := rec.Header().Get("Cache-Control")
				assert.Empty(t, cacheControl, "Cache-Control should not be set for non-static files")
			}
		})
	}
}

func TestShouldCache(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/css/style.css", true},
		{"/js/app.js", true},
		{"/images/logo.png", true},
		{"/images/photo.jpg", true},
		{"/images/graphic.svg", true},
		{"/favicon.ico", true},
		{"/fonts/font.woff2", true},
		{"/manifest.json", true},
		{"/index.html", false},
		{"/api/data", false},
		{"/", false},
		{"/admin/dashboard", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldCache(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionalCacheMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		ifNoneMatch    string
		expectedStatus int
		expectETag     bool
	}{
		{
			name:           "CSS file with matching ETag should return 304",
			path:           "/css/ui.min.css",
			ifNoneMatch:    generateETag("/css/ui.min.css"),
			expectedStatus: http.StatusNotModified,
			expectETag:     true,
		},
		{
			name:           "CSS file with non-matching ETag should return 200",
			path:           "/css/ui.min.css",
			ifNoneMatch:    `"different-etag"`,
			expectedStatus: http.StatusOK,
			expectETag:     false,
		},
		{
			name:           "Non-cacheable file should not return 304",
			path:           "/index.html",
			ifNoneMatch:    `"any-etag"`,
			expectedStatus: http.StatusOK,
			expectETag:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create a simple handler
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "test content")
			}

			// Apply conditional cache middleware
			middleware := conditionalCacheMiddleware()
			err := middleware(handler)(c)

			// Assertions
			if tt.expectedStatus == http.StatusNotModified {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNotModified, rec.Code)
				assert.Empty(t, rec.Body.String())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "test content", rec.Body.String())
			}
		})
	}
}

func TestGenerateETag(t *testing.T) {
	tests := []struct {
		path1    string
		path2    string
		expected bool // should ETags be equal?
	}{
		{"/css/ui.css", "/css/ui.css", true},
		{"/css/ui.css", "/css/other.css", false},
		{"/js/app.js", "/js/app.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.path1+" vs "+tt.path2, func(t *testing.T) {
			etag1 := generateETag(tt.path1)
			etag2 := generateETag(tt.path2)

			assert.NotEmpty(t, etag1)
			assert.NotEmpty(t, etag2)
			assert.True(t, strings.HasPrefix(etag1, `"pg-vis-`))
			assert.True(t, strings.HasPrefix(etag2, `"pg-vis-`))

			if tt.expected {
				assert.Equal(t, etag1, etag2)
			} else {
				assert.NotEqual(t, etag1, etag2)
			}
		})
	}
}

func TestAssetVersion(t *testing.T) {
	// Test that asset version function returns a non-empty string
	version := assetVersion()
	assert.NotEmpty(t, version)

	// Test that multiple calls return the same version
	version2 := assetVersion()
	assert.Equal(t, version, version2)
}

func TestGetAssetVersionParam(t *testing.T) {
	param := getAssetVersionParam()
	assert.NotEmpty(t, param)
	assert.True(t, strings.HasPrefix(param, "v="))
}

func TestMiddlewareWithServerPathPrefix(t *testing.T) {
	// Test with server path prefix
	originalPrefix := serverPathPrefix
	serverPathPrefix = "/pg-press"
	defer func() { serverPathPrefix = originalPrefix }()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pg-press/css/ui.min.css", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test content")
	}

	middleware := staticCacheMiddleware()
	err := middleware(handler)(c)

	assert.NoError(t, err)

	// Should still apply caching even with path prefix
	cacheControl := rec.Header().Get("Cache-Control")
	assert.Equal(t, "public, max-age=31536000, immutable", cacheControl)
}
