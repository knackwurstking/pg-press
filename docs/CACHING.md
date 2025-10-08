# Asset Caching Implementation

This document provides comprehensive documentation for the asset caching strategy implemented in PG Press, covering cache headers, asset versioning, performance optimizations, and monitoring strategies.

## Overview

PG Press implements a multi-layered caching strategy designed to optimize performance while maintaining data freshness. The caching system operates at several levels:

1. **Browser Cache**: Long-term client-side caching for static assets
2. **HTTP Cache Headers**: Server-side cache control directives
3. **Asset Versioning**: Automated cache invalidation through URL versioning
4. **Conditional Requests**: ETags and Last-Modified headers for efficient validation
5. **Application-Level Caching**: Strategic caching of computed data

## Architecture

### Caching Stack

```
┌─────────────────┐
│   Browser       │ ← Long-term asset caching
├─────────────────┤
│   Proxy/CDN     │ ← Edge caching (future)
├─────────────────┤
│   HTTP Headers  │ ← Cache control directives
├─────────────────┤
│   Application   │ ← Dynamic content caching
├─────────────────┤
│   Database      │ ← Query result caching
└─────────────────┘
```

### Core Components

1. **Cache Middleware** (`cmd/pg-press/middleware-cache.go`)
2. **Asset URL Generator** (`internal/web/helpers/assets.go`)
3. **Template Integration** (`.templ` files)
4. **Build System Integration** (`Makefile`)

## Cache Headers Middleware

### Implementation

The cache middleware automatically applies appropriate HTTP headers based on file types and request patterns:

```go
func CacheMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            path := c.Request().URL.Path

            if shouldCache(path) {
                setCacheHeaders(c, path)
            }

            return next(c)
        }
    }
}
```

### Cache Policy by File Type

| File Extension                          | Cache Duration | Cache-Control Header                  | Use Case                          |
| --------------------------------------- | -------------- | ------------------------------------- | --------------------------------- |
| `.css`, `.js`                           | 1 year         | `public, max-age=31536000, immutable` | Versioned stylesheets and scripts |
| `.woff`, `.woff2`, `.ttf`, `.eot`       | 1 year         | `public, max-age=31536000, immutable` | Web fonts                         |
| `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg` | 30 days        | `public, max-age=2592000`             | Images and graphics               |
| `.ico`, `.png` (favicon)                | 1 week         | `public, max-age=604800`              | Favicons and app icons            |
| `.json`, `.xml`                         | 1 day          | `public, max-age=86400`               | Manifest and config files         |
| `.html`                                 | No cache       | `no-cache, must-revalidate`           | Dynamic HTML pages                |
| Other static files                      | 1 hour         | `public, max-age=3600`                | Fallback for unknown types        |

### Cache-Control Directives

**`public`**: Allows caching by browsers and intermediate proxies

- Used for all static assets
- Enables shared caching for better performance

**`max-age`**: Specifies cache duration in seconds

- Eliminates server requests during cache period
- Reduces bandwidth and improves load times

**`immutable`**: Indicates content will never change

- Used for versioned assets (CSS/JS with version parameters)
- Prevents conditional requests during cache period
- Supported by modern browsers for optimal performance

**`no-cache`**: Forces revalidation with server

- Used for HTML pages and dynamic content
- Ensures fresh content while allowing conditional requests

## Asset Versioning System

### Version Generation

Asset versions are generated using multiple strategies:

1. **Environment Variable**: `ASSET_VERSION` (highest priority)
2. **Git Commit Hash**: `GIT_COMMIT` environment variable
3. **Build Timestamp**: Server startup time (fallback)

```go
func generateAssetVersion() string {
    if version := os.Getenv("ASSET_VERSION"); version != "" {
        return version
    }

    if commit := os.Getenv("GIT_COMMIT"); commit != "" {
        return commit[:8] // Use first 8 characters
    }

    return strconv.FormatInt(time.Now().Unix(), 10)
}
```

### URL Generation

The asset URL generator creates versioned URLs for cache busting:

```go
func AssetURL(pathPrefix, assetPath string) string {
    version := getAssetVersion()
    separator := "?"

    if strings.Contains(assetPath, "?") {
        separator = "&"
    }

    return pathPrefix + assetPath + separator + "v=" + version
}
```

### Template Integration

Templates use the asset helper for all static resources:

```html
<!-- CSS -->
<link rel="stylesheet"
      href={ helpers.AssetURL(env.ServerPathPrefix, "/css/ui.min.css") } />

<!-- JavaScript -->
<script src={ helpers.AssetURL(env.ServerPathPrefix, "/js/htmx-v2.0.6.min.js") }>
</script>

<!-- Images -->
<img src={ helpers.AssetURL(env.ServerPathPrefix, "/images/logo.png") }
     alt="Logo" />
```

## Conditional Request Support

### ETag Implementation

ETags are generated for static files to enable efficient conditional requests:

```go
func generateETag(file os.FileInfo) string {
    modTime := file.ModTime().Unix()
    size := file.Size()

    hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d", modTime, size)))
    return fmt.Sprintf("\"%x\"", hash[:8])
}
```

### Last-Modified Headers

Last-Modified headers are set based on file modification times:

```go
func setLastModified(c echo.Context, modTime time.Time) {
    c.Response().Header().Set("Last-Modified",
                              modTime.UTC().Format(http.TimeFormat))
}
```

### Conditional Request Flow

1. Client requests asset with `If-None-Match` or `If-Modified-Since`
2. Server compares ETags or modification times
3. Returns `304 Not Modified` if unchanged
4. Returns full content if changed

## Performance Optimizations

### Compression

Static assets are served with appropriate compression:

```go
// Gzip compression for text-based assets
if isCompressible(contentType) {
    c.Response().Header().Set("Content-Encoding", "gzip")
    // Apply gzip compression
}
```

**Compressible Content Types**:

- `text/css`
- `application/javascript`
- `text/html`
- `application/json`
- `image/svg+xml`

### Content-Type Detection

Accurate Content-Type headers ensure proper browser handling:

```go
func detectContentType(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))

    contentTypes := map[string]string{
        ".css":  "text/css",
        ".js":   "application/javascript",
        ".json": "application/json",
        ".woff": "font/woff",
        ".woff2": "font/woff2",
        // ... more mappings
    }

    return contentTypes[ext]
}
```

### Preload Hints

Critical assets include preload hints for faster loading:

```html
<!-- Critical CSS preload -->
<link rel="preload"
      href={ helpers.AssetURL(env.ServerPathPrefix, "/css/critical.css") }
      as="style" />

<!-- Important JavaScript preload -->
<link rel="preload"
      href={ helpers.AssetURL(env.ServerPathPrefix, "/js/htmx.min.js") }
      as="script" />
```

## Build System Integration

### Development Build

Development builds use timestamp-based versioning for rapid iteration:

```makefile
dev:
	ASSET_VERSION=$(shell date +%s) \
	GIT_COMMIT=$(shell git rev-parse --short HEAD) \
	go run cmd/pg-press/*.go server
```

### Production Build

Production builds use git commits for stable versioning:

```makefile
build-prod:
	GIT_COMMIT=$(shell git rev-parse HEAD) \
	ASSET_VERSION=$(shell git describe --tags --always) \
	go build -ldflags="-w -s" -o bin/pg-press cmd/pg-press/*.go
```

### Asset Pipeline

Assets are processed during build for optimal performance:

```makefile
assets:
	# Minify CSS
	cleancss -o static/css/ui.min.css static/css/ui.css

	# Minify JavaScript
	terser static/js/app.js -o static/js/app.min.js

	# Optimize images
	imagemin static/images/* --out-dir=static/images/optimized
```

## Configuration

### Environment Variables

| Variable                | Description             | Default        | Example           |
| ----------------------- | ----------------------- | -------------- | ----------------- |
| `ASSET_VERSION`         | Override asset version  | auto-generated | `v1.2.3`          |
| `GIT_COMMIT`            | Git commit hash         | auto-detected  | `abc123def456`    |
| `CACHE_CONTROL_MAX_AGE` | Default cache duration  | `3600`         | `86400`           |
| `ENABLE_COMPRESSION`    | Enable gzip compression | `true`         | `false`           |
| `STATIC_FILES_PATH`     | Static files directory  | `./static`     | `/var/www/static` |

### Runtime Configuration

Cache behavior can be configured at runtime:

```go
type CacheConfig struct {
    MaxAge          int           `json:"max_age"`
    EnableETag      bool          `json:"enable_etag"`
    EnableGzip      bool          `json:"enable_gzip"`
    VersionFormat   string        `json:"version_format"`
    StaticFilesPath string        `json:"static_files_path"`
}
```

## Monitoring and Analytics

### Cache Hit Rates

Monitor cache effectiveness through browser developer tools:

1. **Network Tab**: Shows cache status for each request
2. **Performance Tab**: Measures load time improvements
3. **Application Tab**: Shows cached resources

### Key Metrics

**Cache Hit Rate**: Percentage of requests served from cache

- Target: >90% for static assets
- Monitoring: Browser dev tools, server logs

**Load Time Reduction**: Performance improvement from caching

- Target: >50% reduction on repeat visits
- Monitoring: Real User Monitoring (RUM)

**Bandwidth Savings**: Reduced server bandwidth usage

- Target: >70% reduction for return visitors
- Monitoring: Server access logs

### Logging

Cache-related events are logged for monitoring:

```go
logger.Info("Cache hit",
    "path", path,
    "etag", etag,
    "client_etag", clientETag)

logger.Info("Cache miss",
    "path", path,
    "reason", "etag_mismatch")
```

## Debugging and Troubleshooting

### Common Issues

**Assets Not Caching**

1. Check cache headers in browser dev tools
2. Verify middleware is registered correctly
3. Confirm file extensions are in cacheable list

```bash
# Debug cache headers
curl -I http://localhost:8080/css/ui.min.css

# Expected response
HTTP/1.1 200 OK
Cache-Control: public, max-age=31536000, immutable
ETag: "abc123def"
```

**Cache Not Invalidating**

1. Verify asset versioning is working
2. Check that version changes on deployment
3. Clear browser cache manually if needed

```bash
# Check asset version
curl http://localhost:8080/css/ui.min.css?v=123456

# Version should change after deployment
curl http://localhost:8080/css/ui.min.css?v=789012
```

**Performance Issues**

1. Monitor cache hit rates
2. Check if compression is enabled
3. Verify optimal cache durations

```bash
# Test compression
curl -H "Accept-Encoding: gzip" -I http://localhost:8080/js/app.min.js

# Expected response
Content-Encoding: gzip
```

### Debug Tools

**Cache Inspector**: Browser extension for cache analysis
**WebPageTest**: Online performance testing
**Lighthouse**: Chrome DevTools performance audit

```bash
# Local performance testing
lighthouse http://localhost:8080 --only-categories=performance
```

## Security Considerations

### Cache Poisoning Prevention

- Validate all cache keys
- Sanitize file paths
- Implement proper access controls

### Sensitive Data Protection

- Never cache sensitive information
- Use appropriate cache-control for dynamic content
- Implement proper authentication for cached resources

```go
func isCacheable(path string) bool {
    // Never cache sensitive endpoints
    sensitive := []string{"/api/", "/admin/", "/auth/"}

    for _, prefix := range sensitive {
        if strings.HasPrefix(path, prefix) {
            return false
        }
    }

    return isStaticAsset(path)
}
```

## Future Enhancements

### Planned Improvements

1. **Service Worker Caching**
   - Offline support for critical resources
   - Background asset updates
   - Custom caching strategies

2. **CDN Integration**
   - Edge caching for global performance
   - Automatic cache purging
   - Geographic distribution

3. **Advanced Compression**
   - Brotli compression support
   - Content-aware compression
   - Dynamic compression levels

4. **Cache Warming**
   - Automated cache preloading
   - Predictive resource loading
   - Smart prefetching

5. **Real-time Analytics**
   - Cache performance dashboards
   - Automated alerting
   - Performance recommendations

### Implementation Roadmap

**Phase 1**: Service Worker implementation

- Basic offline support
- Critical asset caching
- Update notifications

**Phase 2**: CDN integration

- CloudFront/Cloudflare setup
- Automated deployment
- Edge caching rules

**Phase 3**: Advanced optimization

- Brotli compression
- HTTP/2 push
- Resource hints optimization

## Best Practices

### Development

- Use timestamp-based versions in development
- Test cache behavior across browsers
- Verify cache invalidation on updates

### Production

- Use git-based versioning for stability
- Monitor cache hit rates regularly
- Implement proper backup strategies

### Performance

- Set appropriate cache durations
- Use immutable flag for versioned assets
- Enable compression for all text assets

### Maintenance

- Regular cache performance reviews
- Update caching strategies based on usage patterns
- Keep documentation current with implementation

This comprehensive caching strategy ensures optimal performance while maintaining flexibility for future enhancements and scalability requirements.
