# Asset Caching Implementation

This document describes the asset caching strategy implemented in pg-press to improve performance by leveraging browser caching for static assets.

## Overview

The caching implementation consists of three main components:

1. **Cache Headers Middleware** - Sets appropriate HTTP cache headers for static files
2. **Asset Versioning** - Adds version parameters to asset URLs for cache invalidation
3. **Conditional Request Handling** - Supports ETag and Last-Modified headers for efficient cache validation

## Components

### 1. Cache Headers Middleware (`cmd/pg-press/middleware-cache.go`)

The middleware automatically sets cache headers based on file types:

- **CSS/JS Files**: Cache for 1 year with `immutable` flag
- **Font Files**: Cache for 1 year with `immutable` flag
- **Images**: Cache for 30 days
- **Icons/Favicons**: Cache for 1 week
- **JSON files**: Cache for 1 day
- **Other files**: Cache for 1 hour

### 2. Asset Versioning (`internal/web/helpers/assets.go`)

Provides functions to generate versioned URLs for assets:

```go
// Generate versioned URL
helpers.AssetURL(serverPathPrefix, "/css/ui.min.css")
// Output: /css/ui.min.css?v=abc123ef
```

Version generation:

- Uses server startup timestamp
- Automatically changes on each server restart/deployment
- Simple and reliable cache invalidation

### 3. Template Integration

The main template (`internal/web/templates/layouts/main.templ`) uses versioned asset URLs:

```html
<link rel="stylesheet" href={ helpers.AssetURL(env.ServerPathPrefix, "/css/ui.min.css") }/>
<script src={ helpers.AssetURL(env.ServerPathPrefix, "/js/htmx-v2.0.6.min.js") }></script>
```

## Cache Strategy by File Type

| File Type | Cache Duration | Headers                               | Use Case                            |
| --------- | -------------- | ------------------------------------- | ----------------------------------- |
| CSS/JS    | 1 year         | `public, max-age=31536000, immutable` | Versioned assets that rarely change |
| Fonts     | 1 year         | `public, max-age=31536000, immutable` | Static font files                   |
| Images    | 30 days        | `public, max-age=2592000`             | Regular images                      |
| Icons     | 1 week         | `public, max-age=604800`              | Favicons and app icons              |
| JSON      | 1 day          | `public, max-age=86400`               | Manifest files                      |
| Others    | 1 hour         | `public, max-age=3600`                | Fallback                            |

## Build Integration

### Development

```bash
make dev
# Sets ASSET_VERSION and GIT_COMMIT environment variables
```

### Production Build

```bash
make build-prod
# Generates versioned build with current git commit and timestamp
```

### Manual Version Setting

```bash
ASSET_VERSION=v1.2.3 GIT_COMMIT=abc123ef make run
```

## How It Works

1. **Request Processing**:
   - Middleware checks if the requested file should be cached
   - Sets appropriate cache headers based on file extension
   - Adds ETag and Last-Modified headers for validation

2. **Asset URL Generation**:
   - Template functions generate URLs with version parameters
   - Version changes trigger cache invalidation
   - Browsers fetch new assets when version changes

3. **Browser Behavior**:
   - First visit: Downloads all assets, caches with long expiration
   - Subsequent visits: Serves from cache (no network requests)
   - Version change: Downloads only changed assets

## Benefits

1. **Performance**: Eliminates redundant asset downloads
2. **Bandwidth**: Reduces server bandwidth usage
3. **User Experience**: Faster page loads after initial visit
4. **Scalability**: Reduces server load for static assets

## Cache Invalidation

Cache invalidation happens automatically when:

- Server is restarted
- Application is redeployed
- Server process is reloaded

## Monitoring Cache Effectiveness

Check browser developer tools:

- Network tab shows "from disk cache" or "from memory cache"
- Response headers show cache control directives
- Asset URLs include version parameters

## Environment Variables

| Variable             | Description           | Example     |
| -------------------- | --------------------- | ----------- |
| `SERVER_PATH_PREFIX` | URL prefix for assets | `/pg-press` |

## Troubleshooting

### Assets Not Caching

- Check that middleware is properly registered
- Verify cache headers in browser dev tools
- Ensure file extensions are in the cacheable list

### Cache Not Invalidating

- Verify server has been restarted after deployment
- Check that versioned URLs are being generated with timestamps
- Clear browser cache manually if needed

### Performance Issues

- Monitor cache hit rates
- Check if ETags are being used effectively
- Verify conditional requests are working

## Future Enhancements

1. **Content-Based Hashing**: Use file content hashes instead of timestamps
2. **CDN Integration**: Add support for CDN cache headers
3. **Preloading**: Add resource hints for critical assets
4. **Service Worker**: Implement SW for offline caching
5. **Compression**: Add Brotli/Gzip compression for assets
