// PWA Install Prompt Management
class PWAInstallManager {
    constructor() {
        this.deferredPrompt = null;
        this.init();
    }

    init() {
        // Listen for the beforeinstallprompt event
        window.addEventListener("beforeinstallprompt", (e) => {
            // Prevent the mini-infobar from appearing
            e.preventDefault();
            // Stash the event so it can be triggered later
            this.deferredPrompt = e;
            // Show our custom install prompt
            this.showInstallPrompt();
        });

        // Listen for app installation
        window.addEventListener("appinstalled", () => {
            console.log("[PWA] App was installed");
            this.hideInstallPrompt();
            this.deferredPrompt = null;
        });
    }

    showInstallPrompt() {
        // Don't show if already installed or dismissed recently
        if (this.isInstalled() || this.isRecentlyDismissed()) {
            return;
        }

        const prompt = this.createInstallPrompt();
        document.body.appendChild(prompt);

        // Show with animation
        setTimeout(() => {
            prompt.classList.add("show");
        }, 100);
    }

    createInstallPrompt() {
        const prompt = document.createElement("div");
        prompt.className = "pwa-install-prompt";
        prompt.id = "pwa-install-prompt";

        prompt.innerHTML = `
            <div class="pwa-install-content">
                <div class="pwa-install-text">
                    <div class="pwa-install-title">PG-VIS installieren</div>
                    <div class="pwa-install-message">Zum Startbildschirm hinzufügen für schnellen Zugriff und Offline-Nutzung</div>
                </div>
                <div class="pwa-install-actions">
                    <button class="pwa-install-btn primary" onclick="window.pwaInstall.install()">Installieren</button>
                    <button class="pwa-install-btn" onclick="window.pwaInstall.dismiss()">Später</button>
                </div>
            </div>
        `;

        return prompt;
    }

    async install() {
        if (!this.deferredPrompt) {
            return;
        }

        // Show the install prompt
        this.deferredPrompt.prompt();

        // Wait for the user to respond to the prompt
        const { outcome } = await this.deferredPrompt.userChoice;

        if (outcome === "accepted") {
            console.log("[PWA] User accepted the install prompt");
        } else {
            console.log("[PWA] User dismissed the install prompt");
        }

        this.hideInstallPrompt();
        this.deferredPrompt = null;
    }

    dismiss() {
        this.hideInstallPrompt();
        // Remember dismissal for 7 days
        localStorage.setItem("pwa-install-dismissed", Date.now().toString());
    }

    hideInstallPrompt() {
        const prompt = document.getElementById("pwa-install-prompt");
        if (prompt) {
            prompt.classList.remove("show");
            setTimeout(() => {
                prompt.remove();
            }, 300);
        }
    }

    isInstalled() {
        return (
            window.matchMedia("(display-mode: standalone)").matches ||
            window.navigator.standalone === true
        );
    }

    isRecentlyDismissed() {
        const dismissed = localStorage.getItem("pwa-install-dismissed");
        if (!dismissed) return false;

        const dismissedTime = parseInt(dismissed);
        const sevenDays = 7 * 24 * 60 * 60 * 1000;
        return Date.now() - dismissedTime < sevenDays;
    }
}

// Offline Status Management
class OfflineStatusManager {
    constructor() {
        this.isOnline = navigator.onLine;
        this.init();
    }

    init() {
        // Create offline indicator
        this.createOfflineIndicator();

        // Listen for online/offline events
        window.addEventListener("online", () => this.handleOnline());
        window.addEventListener("offline", () => this.handleOffline());

        // Initial state
        if (!this.isOnline) {
            this.showOfflineIndicator();
        }
    }

    createOfflineIndicator() {
        const indicator = document.createElement("div");
        indicator.className = "offline-indicator";
        indicator.id = "offline-indicator";
        indicator.innerHTML =
            "📡 Sie sind offline - Einige Funktionen sind möglicherweise eingeschränkt";
        document.body.appendChild(indicator);
    }

    handleOnline() {
        this.isOnline = true;
        this.hideOfflineIndicator();
        console.log("[PWA] App is online");

        // Notify service worker
        if (
            "serviceWorker" in navigator &&
            navigator.serviceWorker.controller
        ) {
            navigator.serviceWorker.controller.postMessage({
                type: "ONLINE_STATUS_CHANGED",
                isOnline: true,
            });
        }
    }

    handleOffline() {
        this.isOnline = false;
        this.showOfflineIndicator();
        console.log("[PWA] App is offline");

        // Notify service worker
        if (
            "serviceWorker" in navigator &&
            navigator.serviceWorker.controller
        ) {
            navigator.serviceWorker.controller.postMessage({
                type: "ONLINE_STATUS_CHANGED",
                isOnline: false,
            });
        }
    }

    showOfflineIndicator() {
        const indicator = document.getElementById("offline-indicator");
        if (indicator) {
            indicator.classList.add("show");
        }
    }

    hideOfflineIndicator() {
        const indicator = document.getElementById("offline-indicator");
        if (indicator) {
            indicator.classList.remove("show");
        }
    }
}

// Initialize PWA features when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
    // Initialize PWA install manager
    window.pwaInstall = new PWAInstallManager();

    // Initialize offline status manager
    window.offlineStatus = new OfflineStatusManager();

    // Preload critical pages for offline use
    if (window.swManager) {
        window.swManager.preloadUrls([
            "./trouble-reports",
            "./feed",
            "./profile",
        ]);
    }

    // Add HTMX offline handling
    document.body.addEventListener("htmx:sendError", (event) => {
        console.log("[HTMX] Request failed, possibly offline:", event.detail);

        // Show user-friendly error for offline requests
        if (!navigator.onLine) {
            event.preventDefault();
            // Could show a toast notification here
            console.log("[HTMX] Request failed due to offline status");
        }
    });

    // Log PWA status
    console.log("[PWA] Installation status:", {
        isInstalled: window.pwaInstall.isInstalled(),
        isOnline: navigator.onLine,
        hasServiceWorker: "serviceWorker" in navigator,
    });
});

// Handle page visibility changes for PWA optimization
document.addEventListener("visibilitychange", () => {
    if (!document.hidden) {
        // Page became visible - check for updates
        if (window.swManager) {
            window.swManager.checkForUpdates();
        }

        // Also check websocket connections after suspension
        if (window.webSocketManager) {
            setTimeout(() => {
                window.webSocketManager.checkAndReconnectAll();
            }, 1000);
        }
    }
});

// Keyboard shortcuts for PWA features
document.addEventListener("keydown", (event) => {
    // Ctrl/Cmd + I to show install info
    if ((event.ctrlKey || event.metaKey) && event.key === "i") {
        if (window.pwaInstall && !window.pwaInstall.isInstalled()) {
            event.preventDefault();
            window.pwaInstall.showInstallPrompt();
        }
    }
});
