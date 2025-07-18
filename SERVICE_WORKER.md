# Service Worker Implementation for PG-VIS

This document describes the updated service worker implementation for the PG-VIS (Press Group - Press Visualization) application, providing comprehensive offline functionality and improved user experience.

## Overview

The service worker implementation includes:

- **Intelligent Caching Strategies** - Different strategies for static assets, API endpoints, and pages
- **Offline Support** - Graceful degradation when network is unavailable
- **Background Sync** - Queue actions when offline and sync when online
- **Push Notifications** - Support for real-time updates
- **Cache Management** - Automatic cleanup and version management
- **HTMX Integration** - Optimized for HTMX-based dynamic content

## Architecture

### Cache Structure

The service worker uses multiple cache buckets:

```javascript
const STATIC_CACHE = "pgvis-static-v1.0.0"; // Static assets (CSS, JS, images, fonts)
const DYNAMIC_CACHE = "pgvis-dynamic-v1.0.0"; // API responses and dynamic content
const OFFLINE_CACHE = "pgvis-offline-v1.0.0"; // Offline fallback pages
```

### Caching Strategies

#### 1. Static Assets (Cache First)

- **Files**: CSS, JS, images, fonts, icons
- **Strategy**: Cache first, network fallback
- **Duration**: 7 days
- **Use Case**: Files that rarely change

```javascript
// Static assets pattern
/\.(css|js|woff|woff2|png|jpg|jpeg|gif|svg|ico)$/;
```

#### 2. API Endpoints (Network First)

- **Endpoints**: `/data`, `/dialog-edit`, `/cookies`, `/feed-counter`
- **Strategy**: Network first, cache fallback
- **Duration**: 5 minutes
- **Use Case**: Dynamic content that should be fresh

```javascript
// API endpoints pattern
/\/(data|dialog-edit|cookies|feed-counter)$/;
```

#### 3. Application Pages (Network First with Offline Fallback)

- **Pages**: `/`, `/feed`, `/profile`, `/trouble-reports`
- **Strategy**: Network first, cached fallback, then offline page
- **Duration**: 1 hour
- **Use Case**: Main application pages

#### 4. Authentication (Network Only)

- **Endpoints**: `/login`, `/logout`, `/api-key`
- **Strategy**: Network only, never cached
- **Use Case**: Security-sensitive operations

## File Structure

```
pg-vis/routes/static/
├── service-worker.js          # Main service worker implementation
├── js/sw-register.js         # Client-side registration and management
├── offline.html              # Offline fallback page
└── manifest.json             # PWA manifest (updated)
```

## Features

### 1. Offline Support

When offline, users can:

- Browse cached trouble reports
- View previously loaded feed content
- Access profile information
- Draft new trouble reports (synced when online)

### 2. Background Sync

The service worker queues actions when offline:

```javascript
// Trouble report sync
navigator.serviceWorker.ready.then((registration) => {
    return registration.sync.register("trouble-report-sync");
});

// Profile update sync
navigator.serviceWorker.ready.then((registration) => {
    return registration.sync.register("profile-update-sync");
});
```

### 3. Push Notifications

Support for real-time notifications:

```javascript
// Example notification
{
    title: "PG-VIS Update",
    body: "New trouble report available",
    icon: "./icon.png",
    badge: "./pwa-64x64.png",
    actions: [
        { action: "view", title: "View" },
        { action: "dismiss", title: "Dismiss" }
    ]
}
```

### 4. Smart Cache Management

- **Automatic Cleanup**: Old cache versions are automatically removed
- **Selective Caching**: Only cache successful responses (200 status)
- **Cache Expiration**: Expired entries are automatically removed
- **Storage Monitoring**: Tracks cache usage and performance

## Installation & Setup

### 1. Service Worker Registration

The service worker is automatically registered via `sw-register.js`:

```html
<!-- Include in your HTML templates -->
<script src="./js/sw-register.js"></script>
```

### 2. Manifest Integration

Ensure the manifest is linked in your HTML:

```html
<link rel="manifest" href="./manifest.json" />
```

### 3. HTTPS Requirement

Service workers require HTTPS in production. For development:

- Use `localhost` (works with HTTP)
- Use development certificates
- Use ngrok or similar tunneling service

## Usage Examples

### Preloading Important URLs

```javascript
// Preload critical pages for offline use
window.swManager.preloadUrls(["./trouble-reports", "./feed", "./profile"]);
```

### Checking Cache Status

```javascript
// Get cache information
const status = await window.swManager.getCacheStatus();
console.log("Cache status:", status);
```

### Handling Updates

```javascript
// The service worker manager automatically handles updates
// Users are notified when new versions are available
```

## Offline Page

The offline fallback page (`offline.html`) provides:

- **Network Status Indicator** - Shows current connection state
- **Retry Functionality** - Allows users to test connectivity
- **Feature List** - Shows what's available offline
- **Auto-redirect** - Automatically redirects when back online

### Offline Page Features

- Responsive design for all devices
- Keyboard shortcuts (Ctrl+R to retry, Ctrl+H for home)
- Automatic connection monitoring
- Visual feedback for connection state

## PWA Enhancements

### Manifest Updates

The updated manifest includes:

```json
{
    "name": "PG-VIS - Press Group Visualization",
    "short_name": "PG-VIS",
    "display": "standalone",
    "theme_color": "#667eea",
    "background_color": "#ffffff",
    "shortcuts": [
        {
            "name": "Trouble Reports",
            "url": "./trouble-reports"
        },
        {
            "name": "Feed",
            "url": "./feed"
        },
        {
            "name": "Profile",
            "url": "./profile"
        }
    ]
}
```

### App Shortcuts

Users can access key features directly from:

- Home screen shortcuts
- App launcher context menus
- Desktop shortcuts (when installed)

## Performance Optimizations

### 1. Selective Caching

Only cache resources that provide value offline:

- Static assets for UI
- Recently viewed content
- User-specific data

### 2. Cache Size Management

- Automatic cleanup of old cache versions
- Removal of expired entries
- Monitoring of storage usage

### 3. Network Optimization

- Cache responses to reduce network requests
- Compress cached content when possible
- Use appropriate cache headers

## Browser Support

### Service Worker Support

- ✅ Chrome 45+
- ✅ Firefox 44+
- ✅ Safari 11.1+
- ✅ Edge 17+

### PWA Features

- ✅ Install prompts
- ✅ Offline functionality
- ✅ Background sync (Chrome, Edge)
- ✅ Push notifications
- ✅ App shortcuts

## Troubleshooting

### Common Issues

1. **Service Worker Not Updating**

    ```javascript
    // Force update
    navigator.serviceWorker.getRegistration().then((reg) => {
        reg.update();
    });
    ```

2. **Cache Not Working**
    - Check HTTPS requirement
    - Verify scope configuration
    - Check browser developer tools

3. **Offline Page Not Showing**
    - Ensure offline.html is cached
    - Check network error handling
    - Verify fetch event logic

### Debug Mode

Enable detailed logging:

```javascript
// In service-worker.js
const DEBUG = true; // Set to true for development
```

### Cache Inspection

Use browser developer tools:

1. Open DevTools
2. Go to Application tab
3. Check Cache Storage section
4. Inspect service worker status

## Development Guidelines

### Testing Offline Functionality

1. **Chrome DevTools**:
    - Application tab → Service Workers
    - Check "Offline" checkbox
    - Test app functionality

2. **Network Throttling**:
    - Network tab → Throttling
    - Select "Offline" or "Slow 3G"

3. **Cache Inspection**:
    - Application tab → Cache Storage
    - Inspect cached resources

### Updating the Service Worker

1. Update the `VERSION` constant
2. Modify cache contents if needed
3. Test installation and update flow
4. Deploy and verify auto-update works

### Best Practices

- Always increment version for updates
- Test offline scenarios thoroughly
- Monitor cache storage usage
- Provide meaningful error messages
- Keep offline fallbacks simple and fast

## Security Considerations

### Content Security Policy

Ensure service worker compliance:

```html
<meta
    http-equiv="Content-Security-Policy"
    content="script-src 'self' 'unsafe-inline'; worker-src 'self';"
/>
```

### HTTPS Requirements

- Production deployment must use HTTPS
- Localhost exemption for development only
- Consider certificate automation (Let's Encrypt)

### Data Sensitivity

- Never cache authentication tokens
- Exclude sensitive API endpoints
- Use appropriate cache headers
- Implement cache expiration

## Monitoring & Analytics

### Performance Metrics

Track service worker performance:

```javascript
// Example metrics to monitor
- Cache hit/miss ratios
- Offline usage patterns
- Update success rates
- Error frequencies
```

### User Experience Metrics

- Time to interactive (offline vs online)
- Perceived performance improvements
- Offline feature usage
- User feedback on offline experience

## Migration Guide

### From Previous Version

1. **Clear Old Caches**: The new service worker automatically cleans up old cache versions
2. **Update Registration**: No changes needed if using the registration script
3. **Test Thoroughly**: Verify all features work with new caching strategies

### Breaking Changes

- Cache structure has changed (automatic cleanup handles this)
- Some URL patterns may be cached differently
- Offline fallback behavior has improved

## Future Enhancements

### Planned Features

1. **Advanced Background Sync**
    - Form data persistence
    - Conflict resolution
    - Batch operations

2. **Enhanced Push Notifications**
    - Rich media support
    - Action buttons
    - User preferences

3. **Progressive Loading**
    - Skeleton screens
    - Incremental content loading
    - Predictive caching

4. **Analytics Integration**
    - Offline usage tracking
    - Performance monitoring
    - User behavior analysis

### Contribution Guidelines

When modifying the service worker:

1. Update version number
2. Test offline functionality
3. Update documentation
4. Consider backward compatibility
5. Monitor performance impact

## Support

For issues or questions:

1. Check browser developer tools
2. Review console logs
3. Test in different browsers
4. Verify HTTPS configuration
5. Check service worker registration

## Version History

- **v1.0.0**: Complete rewrite with intelligent caching strategies
- **v0.13**: Previous version with basic caching
- **v0.1**: Initial service worker implementation

---

This service worker implementation provides a robust foundation for offline functionality while maintaining excellent performance and user experience.
