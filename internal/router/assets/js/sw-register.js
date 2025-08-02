/**
 * Service Worker Registration and Management for pgpress
 *
 * This script handles the registration, updates, and communication with the
 * service worker for the pgpress application. It provides user notifications
 * about offline capabilities and handles service worker lifecycle events.
 */

class ServiceWorkerManager {
    constructor() {
        this.swRegistration = null;
        this.isOnline = navigator.onLine;
        this.updateCheckInterval = null;
        this.init();
    }

    /**
     * Initialize the service worker manager
     */
    async init() {
        if (!("serviceWorker" in navigator)) {
            console.warn("[SW Manager] Service Worker not supported");
            return;
        }

        // Register service worker
        await this.registerServiceWorker();

        // Set up event listeners
        this.setupEventListeners();

        // Show offline notification if applicable
        this.showOfflineNotification();
    }

    /**
     * Register the service worker
     */
    async registerServiceWorker() {
        try {
            const registration = await navigator.serviceWorker.register(
                "./service-worker.js",
                {
                    scope: "./",
                },
            );

            this.swRegistration = registration;
            console.log(
                "[SW Manager] Service Worker registered successfully:",
                registration,
            );

            // Handle different registration states
            if (registration.installing) {
                console.log("[SW Manager] Service Worker installing...");
                this.trackInstalling(registration.installing);
            } else if (registration.waiting) {
                console.log("[SW Manager] Service Worker waiting...");
                this.showUpdateNotification();
            } else if (registration.active) {
                console.log("[SW Manager] Service Worker active");
            }

            // Listen for updates
            registration.addEventListener("updatefound", () => {
                console.log("[SW Manager] Service Worker update found");
                this.trackInstalling(registration.installing);
            });
        } catch (error) {
            console.error(
                "[SW Manager] Service Worker registration failed:",
                error,
            );
        }
    }

    /**
     * Track service worker installation progress
     */
    trackInstalling(worker) {
        worker.addEventListener("statechange", () => {
            console.log("[SW Manager] Service Worker state:", worker.state);

            if (worker.state === "installed") {
                if (navigator.serviceWorker.controller) {
                    // New worker available
                    this.showUpdateNotification();
                } else {
                    // First time installation
                    this.showInstallNotification();
                }
            }
        });
    }

    /**
     * Set up event listeners for online/offline and other events
     */
    setupEventListeners() {
        // Online/offline detection
        window.addEventListener("online", () => {
            this.isOnline = true;
            this.hideOfflineNotification();
            console.log("[SW Manager] App is online");
        });

        window.addEventListener("offline", () => {
            this.isOnline = false;
            this.showOfflineNotification();
            console.log("[SW Manager] App is offline");
        });

        // Listen for messages from service worker
        navigator.serviceWorker.addEventListener("message", (event) => {
            this.handleServiceWorkerMessage(event.data);
        });
    }

    /**
     * Handle messages from the service worker
     */
    handleServiceWorkerMessage(data) {
        console.log("[SW Manager] Message from SW:", data);

        switch (data.type) {
            case "CACHE_UPDATED":
                this.showCacheUpdateNotification();
                break;
            case "OFFLINE_READY":
                this.showOfflineReadyNotification();
                break;
            default:
                console.log("[SW Manager] Unknown message type:", data.type);
        }
    }

    /**
     * Check for service worker updates
     */
    async checkForUpdates() {
        if (!this.swRegistration) return;

        console.debug("[SW Manager] Checking for updates...");

        try {
            await this.swRegistration.update();
            console.debug("[SW Manager] Checked for updates");
        } catch (error) {
            console.error("[SW Manager] Update check failed:", error);
        }
    }

    /**
     * Show notification for first-time installation
     */
    showInstallNotification() {
        this.showNotification(
            {
                title: "App Ready for Offline Use",
                message:
                    "pgpress is now available offline! You can use core features even without an internet connection.",
                type: "success",
                persistent: false,
                actions: [
                    {
                        text: "Got it",
                        action: () => this.hideNotification("sw-install"),
                    },
                ],
            },
            "sw-install",
        );
    }

    /**
     * Show notification for app updates
     */
    showUpdateNotification() {
        this.showNotification(
            {
                title: "App Update Available",
                message:
                    "A new version of pgpress is ready. Reload to get the latest features.",
                type: "info",
                persistent: true,
                actions: [
                    {
                        text: "Update Now",
                        action: () => {
                            this.activateUpdate();
                        },
                        primary: true,
                    },
                    {
                        text: "Later",
                        action: () => this.hideNotification("sw-update"),
                    },
                ],
            },
            "sw-update",
        );
    }

    /**
     * Show offline notification
     */
    showOfflineNotification() {
        if (!this.isOnline) {
            this.showNotification(
                {
                    title: "You're Offline",
                    message:
                        "Some features may be limited. Your changes will sync when you're back online.",
                    type: "warning",
                    persistent: true,
                    actions: [
                        {
                            text: "Dismiss",
                            action: () => this.hideNotification("offline"),
                        },
                    ],
                },
                "offline",
            );
        }
    }

    /**
     * Hide offline notification
     */
    hideOfflineNotification() {
        this.hideNotification("offline");
    }

    /**
     * Show cache update notification
     */
    showCacheUpdateNotification() {
        this.showNotification(
            {
                title: "Content Updated",
                message: "New content has been cached for offline use.",
                type: "success",
                persistent: false,
            },
            "cache-update",
        );
    }

    /**
     * Show offline ready notification
     */
    showOfflineReadyNotification() {
        this.showNotification(
            {
                title: "Offline Mode Ready",
                message: "All essential content is now cached for offline use.",
                type: "success",
                persistent: false,
            },
            "offline-ready",
        );
    }

    /**
     * Activate service worker update
     */
    async activateUpdate() {
        console.debug(
            "Activating service worker update",
            this.swRegistration,
            this.swRegistration.waiting,
        );

        if (!this.swRegistration || !this.swRegistration.waiting) {
            this.hideNotification("sw-update");
            window.location.reload();
            return;
        }

        // Tell the waiting service worker to skip waiting
        this.swRegistration.waiting.postMessage({ type: "SKIP_WAITING" });

        window.location.reload();
        this.hideNotification("sw-update");
    }

    /**
     * Generic notification system
     */
    showNotification(config, id) {
        // Remove existing notification with same ID
        this.hideNotification(id);

        const notification = document.createElement("div");
        notification.id = `sw-notification-${id}`;
        notification.className = `sw-notification sw-notification-${config.type}`;

        const typeIcons = {
            success: "✅",
            info: "ℹ️",
            warning: "⚠️",
            error: "❌",
        };

        notification.innerHTML = `
            <div class="sw-notification-content">
                <div class="sw-notification-header">
                    <span class="sw-notification-icon">${typeIcons[config.type] || "ℹ️"}</span>
                    <span class="sw-notification-title">${config.title}</span>
                </div>
                <div class="sw-notification-message">${config.message}</div>
                ${
                    config.actions
                        ? `
                    <div class="sw-notification-actions">
                        ${config.actions
                            .map(
                                (action) => `
                            <button class="sw-notification-btn ${action.primary ? "primary" : ""}"
                                    onclick="window.swManager.executeNotificationAction('${id}', ${config.actions.indexOf(action)})">
                                ${action.text}
                            </button>
                        `,
                            )
                            .join("")}
                    </div>
                `
                        : ""
                }
            </div>
            ${
                !config.persistent
                    ? `
                <button class="sw-notification-close" onclick="window.swManager.hideNotification('${id}')">&times;</button>
            `
                    : ""
            }
        `;

        // Store actions for execution
        notification._actions = config.actions || [];

        document.body.appendChild(notification);

        // Auto-hide non-persistent notifications
        if (!config.persistent) {
            setTimeout(() => {
                this.hideNotification(id);
            }, 5000);
        }
    }

    /**
     * Execute notification action
     */
    executeNotificationAction(notificationId, actionIndex) {
        const notification = document.getElementById(
            `sw-notification-${notificationId}`,
        );
        if (
            notification &&
            notification._actions &&
            notification._actions[actionIndex]
        ) {
            notification._actions[actionIndex].action();
        }
    }

    /**
     * Hide notification
     */
    hideNotification(id) {
        const notification = document.getElementById(`sw-notification-${id}`);
        if (notification) {
            notification.remove();
        }
    }

    /**
     * Preload important URLs for offline use
     */
    preloadUrls(urls) {
        if (!this.swRegistration || !this.swRegistration.active) {
            return;
        }

        this.swRegistration.active.postMessage({
            type: "CACHE_URLS",
            urls: urls,
        });
    }

    /**
     * Get cache status information
     */
    async getCacheStatus() {
        if (!("caches" in window)) {
            return { supported: false };
        }

        try {
            const cacheNames = await caches.keys();
            const pgpressCaches = cacheNames.filter((name) =>
                name.includes("pgpress"),
            );

            let totalSize = 0;
            for (const cacheName of pgpressCaches) {
                const cache = await caches.open(cacheName);
                const keys = await cache.keys();
                totalSize += keys.length;
            }

            return {
                supported: true,
                cacheCount: pgpressCaches.length,
                itemCount: totalSize,
                cacheNames: pgpressCaches,
            };
        } catch (error) {
            console.error("[SW Manager] Cache status check failed:", error);
            return { supported: true, error: error.message };
        }
    }

    /**
     * Clear all caches (for debugging/reset)
     */
    async clearAllCaches() {
        if (!("caches" in window)) {
            return false;
        }

        try {
            const cacheNames = await caches.keys();
            const pgpressCaches = cacheNames.filter((name) =>
                name.includes("pgpress"),
            );

            await Promise.all(
                pgpressCaches.map((cacheName) => caches.delete(cacheName)),
            );

            console.log("[SW Manager] All caches cleared");
            return true;
        } catch (error) {
            console.error("[SW Manager] Cache clearing failed:", error);
            return false;
        }
    }
}

// Initialize service worker manager when DOM is ready
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
        window.swManager = new ServiceWorkerManager();
    });
} else {
    window.swManager = new ServiceWorkerManager();
}

// Add CSS for notifications
const style = document.createElement("style");
style.textContent = `
    .sw-notification {
        position: fixed;
        top: var(--ui-spacing);
        right: var(--ui-spacing);
        max-width: 400px;
        background: var(--ui-bg);
        border-radius: var(--ui-radius);
        box-shadow: 0 8px 24px hsla(var(--ui-hue), var(--ui-saturation), 20%, 0.12),
                    0 4px 8px hsla(var(--ui-hue), var(--ui-saturation), 20%, 0.08);
        border-left: 4px solid var(--ui-primary);
        z-index: 9999;
        animation: slideInRight 0.3s ease-out;
        font-family: var(--ui-font-name);
        border: var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);
        backdrop-filter: blur(5px);
        -webkit-backdrop-filter: blur(5px);
    }

    .sw-notification-success { border-left-color: var(--ui-success); }
    .sw-notification-warning { border-left-color: var(--ui-warning); }
    .sw-notification-error { border-left-color: var(--ui-destructive); }
    .sw-notification-info { border-left-color: var(--ui-info); }

    .sw-notification-content {
        padding: calc(var(--ui-spacing) * 2);
    }

    .sw-notification-header {
        display: flex;
        align-items: center;
        gap: calc(var(--ui-spacing) / 2);
        margin-bottom: calc(var(--ui-spacing) / 2);
    }

    .sw-notification-icon {
        font-size: 1rem;
        color: var(--ui-text);
    }

    .sw-notification-title {
        color: var(--ui-text);
        font-size: 0.875rem;
        font-variation-settings: "MONO" 0, "CASL" 1, "wght" 600, "slnt" -3, "CRSV" 0.5;
    }

    .sw-notification-message {
        color: var(--ui-muted-text);
        font-size: 0.8125rem;
        line-height: var(--ui-line-height);
        margin-bottom: calc(var(--ui-spacing) * 1.5);
        font-variation-settings: "MONO" 0, "CASL" 1, "wght" 400, "slnt" 0, "CRSV" 0.5;
    }

    .sw-notification-actions {
        display: flex;
        gap: calc(var(--ui-spacing) / 2);
        flex-wrap: wrap;
    }

    .sw-notification-btn {
        padding: calc(var(--ui-spacing) / 2) var(--ui-spacing);
        border: var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);
        background: var(--ui-bg);
        border-radius: var(--ui-radius);
        font-size: 0.75rem;
        cursor: pointer;
        transition: all 0.25s ease-in-out;
        color: var(--ui-text);
        font-variation-settings: "MONO" 0, "CASL" 1, "wght" 500, "slnt" 0, "CRSV" 0.5;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        text-transform: capitalize;
        user-select: none;
        -webkit-user-select: none;
        -moz-user-select: none;
    }

    .sw-notification-btn:hover {
        background: var(--ui-muted);
        transform: scale(1.02);
    }

    .sw-notification-btn:active {
        transform: scale(0.98);
        transition: none;
    }

    .sw-notification-btn.primary {
        background: var(--ui-primary);
        color: var(--ui-primary-text);
        border-color: var(--ui-primary);
    }

    .sw-notification-btn.primary:hover {
        background: var(--ui-primary--hover);
        border-color: var(--ui-primary--hover);
    }

    .sw-notification-btn.primary:active {
        background: var(--ui-primary--active);
        border-color: var(--ui-primary--active);
    }

    .sw-notification-close {
        position: absolute;
        top: calc(var(--ui-spacing) / 2);
        right: calc(var(--ui-spacing) / 2);
        background: none;
        border: none;
        font-size: 1.125rem;
        cursor: pointer;
        color: var(--ui-muted-text);
        width: 1.5rem;
        height: 1.5rem;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--ui-radius);
        transition: all 0.2s ease;
        user-select: none;
        -webkit-user-select: none;
        -moz-user-select: none;
    }

    .sw-notification-close:hover {
        background: var(--ui-muted);
        color: var(--ui-text);
        transform: scale(1.1);
    }

    .sw-notification-close:active {
        transform: scale(0.9);
        transition: none;
    }

    .sw-notification-close:focus-visible {
        outline: 2px solid var(--ui-primary);
        outline-offset: 2px;
    }

    @keyframes slideInRight {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    @media (max-width: 480px) {
        .sw-notification {
            left: var(--ui-spacing);
            right: var(--ui-spacing);
            max-width: none;
        }
    }
`;
document.head.appendChild(style);
