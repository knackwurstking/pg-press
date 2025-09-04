# Example Output: Versioned Assets in HTML

This document shows examples of how the asset caching implementation generates versioned URLs in the final HTML output.

## Before Implementation

Without asset versioning, the HTML would look like this:

```html
<!DOCTYPE html>
<html lang="de">
    <head>
        <!-- Stylesheets -->
        <link rel="stylesheet" href="/pg-press/css/bootstrap-icons.min.css" />
        <link rel="stylesheet" href="/pg-press/css/ui.min.css" />
        <link rel="stylesheet" href="/pg-press/css/layout.css" />

        <!-- HTMX for dynamic content -->
        <script src="/pg-press/js/htmx-v2.0.6.min.js"></script>
        <script src="/pg-press/js/htmx-ext-ws-v2.0.3.min.js"></script>

        <!-- Icons -->
        <link rel="icon" href="/pg-press/favicon.ico" sizes="any" />
        <link rel="icon" href="/pg-press/icon.png" type="image/png" />
    </head>
</html>
```

## After Implementation

With asset versioning enabled, the same HTML is generated with version parameters:

```html
<!DOCTYPE html>
<html lang="de">
    <head>
        <!-- Stylesheets with versioning -->
        <link
            rel="stylesheet"
            href="/pg-press/css/bootstrap-icons.min.css?v=1705405800"
        />
        <link rel="stylesheet" href="/pg-press/css/ui.min.css?v=1705405800" />
        <link rel="stylesheet" href="/pg-press/css/layout.css?v=1705405800" />

        <!-- HTMX with versioning -->
        <script src="/pg-press/js/htmx-v2.0.6.min.js?v=1705405800"></script>
        <script src="/pg-press/js/htmx-ext-ws-v2.0.3.min.js?v=1705405800"></script>

        <!-- Icons with versioning -->
        <link
            rel="icon"
            href="/pg-press/favicon.ico?v=1705405800"
            sizes="any"
        />
        <link
            rel="icon"
            href="/pg-press/icon.png?v=1705405800"
            type="image/png"
        />
    </head>
</html>
```

## HTTP Response Headers

When a browser requests `/pg-press/css/ui.min.css?v=1705405800`, the server responds with:

```http
HTTP/1.1 200 OK
Cache-Control: public, max-age=31536000, immutable
Expires: Wed, 15 Jan 2025 10:30:00 GMT
ETag: "pg-vis-a1b2c3d4"
Last-Modified: Tue, 15 Jan 2024 10:30:00 GMT
Vary: Accept-Encoding
Content-Type: text/css; charset=utf-8
Content-Length: 15234

/* CSS content here */
```

## Browser Cache Behavior

### First Visit

1. Browser requests all assets with version parameters
2. Server responds with long cache headers
3. Browser caches assets locally
4. Page loads in ~500ms

### Subsequent Visits (Same Version)

1. Browser checks cache for each asset
2. Finds cached versions with matching URLs
3. Serves from cache (no network requests)
4. Page loads in ~50ms

### After Deployment (New Version)

1. HTML now contains `?v=1705408000` instead of `?v=1705405800`
2. Browser treats these as new URLs
3. Requests fresh assets from server
4. Old cached assets eventually expire

## Cache Validation Example

When a browser has an asset cached and wants to check if it's still valid:

### Browser Request

```http
GET /pg-press/css/ui.min.css?v=1705405800 HTTP/1.1
Host: localhost:9020
If-None-Match: "pg-vis-a1b2c3d4"
If-Modified-Since: Tue, 15 Jan 2024 10:30:00 GMT
```

### Server Response (Asset Unchanged)

```http
HTTP/1.1 304 Not Modified
ETag: "pg-vis-a1b2c3d4"
Cache-Control: public, max-age=31536000, immutable
```

## Version Sources

The version parameter (`v=1705405800`) is generated from the server startup time:

### Production Build

```bash
make build-prod
# Uses startup timestamp: v=1705405800
```

### Development

```bash
make dev
# Uses startup timestamp: v=1705405800
```

### How It Works

The version changes automatically every time the server starts, ensuring cache invalidation on deployments without requiring environment variables or git integration.

## Browser Developer Tools

In Chrome/Firefox developer tools, you'll see:

### Network Tab (First Load)

```
Name                                          Status  Type    Size      Time
css/ui.min.css?v=1705405800                  200     css     15.2 KB   45ms
css/bootstrap-icons.min.css?v=1705405800     200     css     8.7 KB    32ms
js/htmx-v2.0.6.min.js?v=1705405800          200     js      12.1 KB   38ms
```

### Network Tab (Cached Load)

```
Name                                          Status     Type    Size      Time
css/ui.min.css?v=1705405800                  from cache css     15.2 KB   0ms
css/bootstrap-icons.min.css?v=1705405800     from cache css     8.7 KB    0ms
js/htmx-v2.0.6.min.js?v=1705405800          from cache js      12.1 KB   0ms
```

## Performance Impact

### Without Caching

- Every page load: ~250KB downloaded
- Load time: ~500ms
- Server requests: 15+ per page

### With Caching

- First load: ~250KB downloaded, then cached
- Subsequent loads: ~2KB (HTML only)
- Load time: ~50ms
- Server requests: 1 per page (HTML only)

### Cache Hit Rate

- Expected cache hit rate: >95% for returning users
- Bandwidth savings: ~248KB per page view
- Server load reduction: ~90% for static assets

## Integration with CDN

When using a CDN, the versioned URLs work seamlessly:

```html
<!-- CDN URLs with versioning -->
<link
    rel="stylesheet"
    href="https://cdn.example.com/pg-press/css/ui.min.css?v=1705405800"
/>
<script src="https://cdn.example.com/pg-press/js/htmx-v2.0.6.min.js?v=1705405800"></script>
```

The CDN will cache these as separate resources, ensuring proper cache invalidation when versions change.
