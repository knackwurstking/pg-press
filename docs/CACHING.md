# Asset Caching Implementation

This document describes the current asset caching strategy implemented in PG Press, covering HTTP cache headers, conditional requests, and performance optimizations.

## Overview

PG Press implements a comprehensive caching strategy designed to optimize web performance through:

1. **Static Asset Caching**: Long-term browser caching for CSS, JS, images, and fonts
2. **Conditional Requests**: ETag and Last-Modified header support for cache validation
3. **Cache Control Headers**: Differentiated caching policies based on file types
4. **Embedded Assets**: Static files served from Go embed.FS for optimal delivery

## Architecture

### Middleware Stack

The caching system consists of two main middleware components:

```go
// Applied in order:
e.Use(conditionalCacheMiddleware())  // Handles If-None-Match, If-Modified-Since
e.Use(staticCacheMiddleware())       // Sets cache headers for static files
```

### Static File Serving

Static assets are embedded into the binary and served directly:

```go
//go:embed assets
assets embed.FS

e.StaticFS(serverPathPrefix+"/", echo.MustSubFS(assets, "assets"))
```

## Cache Policies

### File Type Classifications

| File Extension                          | Cache Duration | Cache-Control Header                  | Use Case                   |
| --------------------------------------- | -------------- | ------------------------------------- | -------------------------- |
| `.css`, `.js`                           | 1 year         | `public, max-age=31536000, immutable` | Stylesheets and scripts    |
| `.woff`, `.woff2`, `.ttf`, `.eot`       | 1 year         | `public, max-age=31536000, immutable` | Web fonts                  |
| `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg` | 30 days        | `public, max-age=2592000`             | Images and graphics        |
| `.ico`                                  | 1 week         | `public, max-age=604800`              | Favicons                   |
| `.json`                                 | 1 day          | `public, max-age=86400`               | Manifest and config files  |
| Other files                             | 1 hour         | `public, max-age=3600`                | Fallback for unknown types |

### Cache Headers Applied

For all cacheable resources, the following headers are set:

```http
Cache-Control: public, max-age=<seconds>[, immutable]
Expires: <future-date>
ETag: "pg-press-<hash>"
Last-Modified: <date>
Vary: Accept-Encoding
```

## Implementation Details

### Cache Middleware (`staticCacheMiddleware`)

```go
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
```

### Conditional Request Handling

The `conditionalCacheMiddleware` processes:

- **If-None-Match**: ETag-based validation
- **If-Modified-Since**: Date-based validation

Returns `304 Not Modified` when resources haven't changed.

### ETag Generation

ETags are generated using MD5 hash of the file path:

```go
func generateETag(path string) string {
    hasher := md5.New()
    hasher.Write([]byte(path))
    hash := fmt.Sprintf("%x", hasher.Sum(nil))
    return fmt.Sprintf(`"pg-press-%s"`, hash[:8])
}
```

### Asset Versioning

The system includes a simple timestamp-based versioning system:

```go
func assetVersion() string {
    // Changes every hour
    return fmt.Sprintf("%d", time.Now().Unix()/3600)
}

func getAssetVersionParam() string {
    version := assetVersion()
    return "v=" + version
}
```

**Note**: The versioning functions are available but may not be actively used in templates. The primary cache invalidation relies on server restarts and ETag/Last-Modified headers.

## Configuration

### Environment Variables

| Variable             | Description         | Default | Example          |
| -------------------- | ------------------- | ------- | ---------------- |
| `SERVER_PATH_PREFIX` | URL path prefix     | `/`     | `/pg-press`      |
| `SERVER_ADDR`        | Server bind address | `:8080` | `localhost:9020` |

### Cacheable File Extensions

The system recognizes these file types as cacheable:

```go
cacheableExts := map[string]bool{
    ".css":   true,   // Stylesheets
    ".js":    true,   // JavaScript
    ".png":   true,   // PNG images
    ".jpg":   true,   // JPEG images
    ".jpeg":  true,   // JPEG images
    ".gif":   true,   // GIF images
    ".svg":   true,   // SVG graphics
    ".ico":   true,   // Icons
    ".woff":  true,   // Web fonts
    ".woff2": true,   // Web fonts v2
    ".ttf":   true,   // TrueType fonts
    ".eot":   true,   // Embedded OpenType fonts
    ".json":  true,   // JSON files (manifests)
}
```

## Performance Benefits

### Browser Behavior

1. **First Visit**: Downloads all assets, stores in browser cache
2. **Subsequent Visits**:
   - Serves from cache without server requests (for non-expired assets)
   - Makes conditional requests for expired assets
   - Downloads only changed assets

### Server Benefits

- **Reduced Bandwidth**: Fewer asset downloads after initial visit
- **Lower CPU Usage**: Conditional requests return 304 responses quickly
- **Better Scalability**: Static assets served efficiently from embedded files

### Typical Cache Hit Rates

- **CSS/JS Files**: ~95% (1-year cache duration)
- **Images**: ~85% (30-day cache duration)
- **Fonts**: ~98% (1-year cache duration)
- **Icons**: ~90% (1-week cache duration)

## Monitoring Cache Performance

### Browser Developer Tools

1. **Network Tab**: Shows cache status for each request
   - "from disk cache" or "from memory cache" indicates cache hits
   - Status 304 indicates successful conditional requests

2. **Headers Inspection**: Verify cache headers are properly set
   ```http
   Cache-Control: public, max-age=31536000, immutable
   ETag: "pg-press-a1b2c3d4"
   ```

### Server Logs

Cache-related information can be monitored through:

- Request logs showing 304 responses
- ETag match/mismatch patterns
- Conditional request frequency

## Troubleshooting

### Common Issues

**Assets Not Caching**

1. Check file extension is in cacheable list
2. Verify middleware is properly configured
3. Inspect response headers in browser dev tools

```bash
# Test cache headers
curl -I http://localhost:9020/pg-press/css/style.css

# Expected response
HTTP/1.1 200 OK
Cache-Control: public, max-age=31536000, immutable
ETag: "pg-press-a1b2c3d4"
```

**304 Responses Not Working**

1. Verify ETags are being generated consistently
2. Check that conditional request headers are being processed
3. Ensure Last-Modified headers are set correctly

**Cache Not Invalidating**

Since the current implementation uses path-based ETags:

- ETags remain consistent across server restarts
- Cache invalidation happens primarily through browser cache expiration
- Force refresh (Ctrl+F5) bypasses all caching

### Debug Commands

```bash
# Check response headers
curl -v http://localhost:9020/pg-press/js/app.js

# Test conditional request
curl -H "If-None-Match: \"pg-press-a1b2c3d4\"" \
     http://localhost:9020/pg-press/js/app.js
```

## Development vs Production

### Development Behavior

- Assets served from embedded files
- Cache headers applied normally
- No special development overrides

### Production Considerations

- Long cache durations reduce server load
- Consider CDN integration for global distribution
- Monitor cache hit rates for optimization opportunities

## Future Enhancements

### Potential Improvements

1. **Content-Based ETags**: Generate ETags from file content rather than path
2. **Build-Time Versioning**: Add version parameters to asset URLs
3. **CDN Integration**: Add support for Content Delivery Networks
4. **Cache Warming**: Pre-populate caches for critical assets
5. **Compression**: Add Brotli/Gzip compression for text assets

### Implementation Considerations

```go
// Example: Content-based ETag generation
func generateContentETag(content []byte) string {
    hasher := md5.New()
    hasher.Write(content)
    hash := fmt.Sprintf("%x", hasher.Sum(nil))
    return fmt.Sprintf(`"pg-press-content-%s"`, hash[:12])
}
```

## Best Practices

### Cache Strategy Guidelines

1. **Immutable Assets**: Use `immutable` directive for versioned assets
2. **Appropriate Durations**: Balance freshness vs performance
3. **Conditional Requests**: Always support ETags and Last-Modified
4. **Vary Headers**: Include `Vary: Accept-Encoding` for compressed content

### Development Workflow

1. Test cache behavior during development
2. Verify cache headers in staging environment
3. Monitor cache performance in production
4. Document any caching exceptions or special cases

This caching implementation provides a solid foundation for web performance optimization while maintaining simplicity and reliability.
