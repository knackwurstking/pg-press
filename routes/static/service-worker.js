/**
 * Service Worker for PG-VIS
 *
 * This service worker provides offline functionality and caching for the PG-VIS
 * (Press Group - Press Visualization) web application. It implements a cache-first
 * strategy for static assets and a network-first strategy for dynamic content.
 *
 * Features:
 * - Offline support for static assets (CSS, JS, images, fonts)
 * - Cache management with automatic cleanup of old versions
 * - Background sync for forms when offline
 * - Push notification support for new trouble reports and feeds
 * - HTMX-aware caching for dynamic content updates
 *
 * Cache Strategy:
 * - Static assets: Cache first, network fallback
 * - API endpoints: Network first, cache fallback
 * - Pages: Network first with offline fallback
 */

// Version and cache configuration
const VERSION = "v0.0.1";
const CACHE_PREFIX = "pgvis";
const STATIC_CACHE = `${CACHE_PREFIX}-static-${VERSION}`;
const DYNAMIC_CACHE = `${CACHE_PREFIX}-dynamic-${VERSION}`;
const OFFLINE_CACHE = `${CACHE_PREFIX}-offline-${VERSION}`;

// Cache duration settings (in milliseconds)
const CACHE_DURATION = {
    STATIC: 7 * 24 * 60 * 60 * 1000, // 7 days
    DYNAMIC: 1 * 60 * 60 * 1000, // 1 hour
    API: 5 * 60 * 1000, // 5 minutes
};

// Static assets to cache on install
const STATIC_ASSETS = [
    // Core application files
    "./",
    "./manifest.json",
    "./favicon.ico",
    "./icon.png",
    "./offline.html",

    // PWA icons
    "./apple-touch-icon-180x180.png",
    "./pwa-64x64.png",
    "./pwa-192x192.png",
    "./pwa-512x512.png",
    "./maskable-icon-512x512.png",

    // Stylesheets
    "./ui-dev.min.css",
    "./css/bootstrap-icons.min.css",

    // Fonts
    "./bootstrap-icons.woff",
    "./bootstrap-icons.woff2",

    // JavaScript libraries
    "./htmx.min.js",
];

// Offline fallback pages
const OFFLINE_FALLBACKS = {
    page: "./offline.html",
    image: "./icon.png",
};

// URL patterns for different caching strategies
const URL_PATTERNS = {
    // Static assets that should be cached first
    static: /\.(css|js|woff|woff2|png|jpg|jpeg|gif|svg|ico)$/,

    // API endpoints that should use network-first
    api: /\/(data|dialog-edit|cookies|feed-counter)$/,

    // HTMX partial updates
    htmx: /\/(data|dialog-edit|cookies|nav\/feed-counter)$/,

    // Authentication related endpoints (never cache)
    auth: /\/(login|logout|api-key)$/,

    // Main application pages
    pages: /\/(feed|profile|trouble-reports)$/,
};

/**
 * Service Worker Installation
 *
 * Pre-caches essential static assets and sets up the foundation
 * for offline functionality.
 */
self.addEventListener("install", (event) => {
    console.log(`[SW] Installing version ${VERSION}`);

    // Skip waiting to activate immediately
    self.skipWaiting();

    event.waitUntil(
        caches
            .open(STATIC_CACHE)
            .then((cache) => {
                console.log("[SW] Caching static assets");
                return cache.addAll(STATIC_ASSETS);
            })
            .then(() => {
                console.log("[SW] Static assets cached successfully");
            })
            .catch((error) => {
                console.error("[SW] Failed to cache static assets:", error);
            }),
    );
});

/**
 * Service Worker Activation
 *
 * Cleans up old caches and claims all clients to ensure the new
 * service worker takes control immediately.
 */
self.addEventListener("activate", (event) => {
    console.log(`[SW] Activating version ${VERSION}`);

    event.waitUntil(
        Promise.all([
            // Clean up old caches
            caches.keys().then((cacheNames) => {
                return Promise.all(
                    cacheNames
                        .filter((cacheName) => {
                            return (
                                cacheName.startsWith(CACHE_PREFIX) &&
                                !cacheName.includes(VERSION)
                            );
                        })
                        .map((cacheName) => {
                            console.log(
                                `[SW] Deleting old cache: ${cacheName}`,
                            );
                            return caches.delete(cacheName);
                        }),
                );
            }),

            // Claim all clients
            self.clients.claim(),
        ]),
    );

    console.log("[SW] Activation complete");
});

/**
 * Fetch Event Handler
 *
 * Implements intelligent caching strategies based on request type:
 * - Static assets: Cache first with network fallback
 * - API endpoints: Network first with cache fallback
 * - Pages: Network first with offline fallback
 * - Auth endpoints: Network only (never cached)
 */
self.addEventListener("fetch", (event) => {
    const { request } = event;
    const { url, method } = request;

    // Only handle GET requests
    if (method !== "GET") {
        return;
    }

    // Skip chrome-extension and other non-http requests
    if (!url.startsWith("http")) {
        return;
    }

    event.respondWith(handleFetch(request));
});

/**
 * Handles fetch requests with appropriate caching strategy
 */
async function handleFetch(request) {
    const url = new URL(request.url);
    const pathname = url.pathname;

    try {
        // Authentication endpoints - network only, never cache
        if (URL_PATTERNS.auth.test(pathname)) {
            return await fetch(request);
        }

        // Static assets - cache first
        if (URL_PATTERNS.static.test(pathname)) {
            return await cacheFirst(request, STATIC_CACHE);
        }

        // API endpoints and HTMX partials - network first
        if (
            URL_PATTERNS.api.test(pathname) ||
            URL_PATTERNS.htmx.test(pathname)
        ) {
            return await networkFirst(
                request,
                DYNAMIC_CACHE,
                CACHE_DURATION.API,
            );
        }

        // Main application pages - network first with offline fallback
        if (URL_PATTERNS.pages.test(pathname) || pathname === "/") {
            return await networkFirstWithOfflineFallback(request);
        }

        // Default: network first for other requests
        return await networkFirst(
            request,
            DYNAMIC_CACHE,
            CACHE_DURATION.DYNAMIC,
        );
    } catch (error) {
        console.error("[SW] Fetch error:", error);

        // Return offline fallback for navigation requests
        if (request.mode === "navigate") {
            return await getOfflineFallback("page");
        }

        // Return generic offline response
        return new Response("Offline", {
            status: 503,
            statusText: "Service Unavailable",
        });
    }
}

/**
 * Cache First Strategy
 *
 * Tries cache first, falls back to network if not found.
 * Ideal for static assets that don't change frequently.
 */
async function cacheFirst(request, cacheName) {
    const cache = await caches.open(cacheName);
    const cachedResponse = await cache.match(request);

    if (cachedResponse) {
        console.log(`[SW] Cache hit: ${request.url}`);
        return cachedResponse;
    }

    console.log(`[SW] Cache miss, fetching: ${request.url}`);
    const networkResponse = await fetch(request);

    // Cache successful responses
    if (networkResponse.ok) {
        cache.put(request, networkResponse.clone());
    }

    return networkResponse;
}

/**
 * Network First Strategy
 *
 * Tries network first, falls back to cache if network fails.
 * Ideal for dynamic content that should be fresh when possible.
 */
async function networkFirst(
    request,
    cacheName,
    maxAge = CACHE_DURATION.DYNAMIC,
) {
    const cache = await caches.open(cacheName);

    try {
        const networkResponse = await fetch(request);

        if (networkResponse.ok) {
            // Add timestamp for cache expiration
            const responseToCache = networkResponse.clone();
            const headers = new Headers(responseToCache.headers);
            headers.set("sw-cached-at", Date.now().toString());

            const modifiedResponse = new Response(responseToCache.body, {
                status: responseToCache.status,
                statusText: responseToCache.statusText,
                headers: headers,
            });

            cache.put(request, modifiedResponse);
            console.log(`[SW] Network success, cached: ${request.url}`);
        }

        return networkResponse;
    } catch (error) {
        console.log(`[SW] Network failed, trying cache: ${request.url}`);

        const cachedResponse = await cache.match(request);

        if (cachedResponse) {
            // Check if cache is still valid
            const cachedAt = cachedResponse.headers.get("sw-cached-at");
            if (cachedAt && Date.now() - parseInt(cachedAt) < maxAge) {
                console.log(`[SW] Cache hit (valid): ${request.url}`);
                return cachedResponse;
            } else {
                console.log(`[SW] Cache expired: ${request.url}`);
                // Remove expired cache entry
                cache.delete(request);
            }
        }

        throw error;
    }
}

/**
 * Network First with Offline Fallback
 *
 * Specialized strategy for navigation requests that provides
 * a meaningful offline page when network fails.
 */
async function networkFirstWithOfflineFallback(request) {
    try {
        const networkResponse = await fetch(request);

        if (networkResponse.ok) {
            // Cache successful page responses
            const cache = await caches.open(DYNAMIC_CACHE);
            cache.put(request, networkResponse.clone());
        }

        return networkResponse;
    } catch (error) {
        console.log(`[SW] Network failed for page: ${request.url}`);

        // Try cached version first
        const cache = await caches.open(DYNAMIC_CACHE);
        const cachedResponse = await cache.match(request);

        if (cachedResponse) {
            return cachedResponse;
        }

        // Return offline fallback page
        return await getOfflineFallback("page");
    }
}

/**
 * Returns appropriate offline fallback content
 */
async function getOfflineFallback(type) {
    const cache = await caches.open(OFFLINE_CACHE);
    const fallbackUrl = OFFLINE_FALLBACKS[type];

    if (fallbackUrl) {
        const fallback = await cache.match(fallbackUrl);
        if (fallback) {
            return fallback;
        }
    }

    // Redirect to proper offline page with CSS variables
    return Response.redirect("./offline.html", 302);
}

/**
 * Background Sync Event Handler
 *
 * Handles queued actions when the device comes back online.
 * Useful for form submissions that failed due to network issues.
 */
self.addEventListener("sync", (event) => {
    console.log("[SW] Background sync triggered:", event.tag);

    if (event.tag === "trouble-report-sync") {
        event.waitUntil(syncTroubleReports());
    } else if (event.tag === "profile-update-sync") {
        event.waitUntil(syncProfileUpdates());
    }
});

/**
 * Push Event Handler
 *
 * Handles push notifications for new content updates.
 */
self.addEventListener("push", (event) => {
    console.log("[SW] Push message received");

    const options = {
        body: "New content available",
        icon: "./icon.png",
        badge: "./pwa-64x64.png",
        vibrate: [100, 50, 100],
        data: {
            dateOfArrival: Date.now(),
            primaryKey: 1,
        },
        actions: [
            {
                action: "explore",
                title: "View Updates",
                icon: "./icon.png",
            },
            {
                action: "close",
                title: "Close",
                icon: "./icon.png",
            },
        ],
    };

    event.waitUntil(
        self.registration.showNotification("PG-VIS Update", options),
    );
});

/**
 * Notification Click Handler
 */
self.addEventListener("notificationclick", (event) => {
    console.log("[SW] Notification clicked:", event.action);

    event.notification.close();

    if (event.action === "explore") {
        event.waitUntil(clients.openWindow("./"));
    }
});

/**
 * Message Event Handler
 *
 * Handles messages from the main application thread.
 */
self.addEventListener("message", (event) => {
    console.log("[SW] Message received:", event.data);

    if (event.data && event.data.type === "SKIP_WAITING") {
        self.skipWaiting();
    } else if (event.data && event.data.type === "CACHE_URLS") {
        event.waitUntil(cacheUrls(event.data.urls));
    }
});

/**
 * Sync functions for background operations
 */
async function syncTroubleReports() {
    // Implementation for syncing queued trouble reports
    console.log("[SW] Syncing trouble reports...");
    // This would integrate with IndexedDB to process queued submissions
}

async function syncProfileUpdates() {
    // Implementation for syncing queued profile updates
    console.log("[SW] Syncing profile updates...");
    // This would integrate with IndexedDB to process queued updates
}

/**
 * Cache additional URLs on demand
 */
async function cacheUrls(urls) {
    const cache = await caches.open(DYNAMIC_CACHE);

    for (const url of urls) {
        try {
            const response = await fetch(url);
            if (response.ok) {
                await cache.put(url, response);
                console.log(`[SW] Cached URL: ${url}`);
            }
        } catch (error) {
            console.error(`[SW] Failed to cache URL: ${url}`, error);
        }
    }
}

// Log service worker status
console.log(`[SW] Service Worker loaded - Version: ${VERSION}`);
