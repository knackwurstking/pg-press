// WebSocket Connection Manager for Browser Suspension Handling
class WebSocketManager {
    constructor() {
        this.connections = new Map();
        this.reconnectAttempts = new Map();
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start with 1 second
        this.maxReconnectDelay = 30000; // Max 30 seconds
        this.isPageVisible = !document.hidden;
        this.init();
    }

    init() {
        // Listen for page visibility changes
        document.addEventListener('visibilitychange', () => {
            this.handleVisibilityChange();
        });

        // Listen for online/offline events
        window.addEventListener('online', () => {
            this.handleOnline();
        });

        window.addEventListener('offline', () => {
            this.handleOffline();
        });

        // Monitor HTMX websocket events
        this.setupHTMXWebSocketMonitoring();

        console.log('[WebSocketManager] Initialized websocket suspension handling');
    }

    setupHTMXWebSocketMonitoring() {
        // Monitor HTMX websocket connections
        document.body.addEventListener('htmx:wsOpen', (event) => {
            const element = event.target;
            const url = this.getWebSocketUrl(element);
            if (url) {
                this.registerConnection(url, element);
                console.log('[WebSocketManager] WebSocket opened:', url);
            }
        });

        document.body.addEventListener('htmx:wsClose', (event) => {
            const element = event.target;
            const url = this.getWebSocketUrl(element);
            if (url) {
                console.log('[WebSocketManager] WebSocket closed:', url);
                // Don't immediately reconnect if page is hidden (suspended)
                if (this.isPageVisible && navigator.onLine) {
                    this.scheduleReconnection(url, element);
                }
            }
        });

        document.body.addEventListener('htmx:wsError', (event) => {
            const element = event.target;
            const url = this.getWebSocketUrl(element);
            if (url) {
                console.log('[WebSocketManager] WebSocket error:', url, event.detail);
                // Only reconnect if page is visible
                if (this.isPageVisible && navigator.onLine) {
                    this.scheduleReconnection(url, element);
                }
            }
        });
    }

    getWebSocketUrl(element) {
        const wsAttribute = element.getAttribute('hx-ws');
        if (wsAttribute && wsAttribute.startsWith('connect:')) {
            return wsAttribute.replace('connect:', '');
        }
        return null;
    }

    registerConnection(url, element) {
        this.connections.set(url, {
            element: element,
            lastConnected: Date.now(),
            isConnected: true
        });

        // Reset reconnect attempts on successful connection
        this.reconnectAttempts.set(url, 0);
    }

    scheduleReconnection(url, element) {
        const attempts = this.reconnectAttempts.get(url) || 0;

        if (attempts >= this.maxReconnectAttempts) {
            console.log(`[WebSocketManager] Max reconnect attempts reached for ${url}`);
            return;
        }

        // Calculate exponential backoff delay
        const delay = Math.min(
            this.reconnectDelay * Math.pow(2, attempts),
            this.maxReconnectDelay
        );

        this.reconnectAttempts.set(url, attempts + 1);

        console.log(`[WebSocketManager] Scheduling reconnection for ${url} in ${delay}ms (attempt ${attempts + 1})`);

        setTimeout(() => {
            // Only reconnect if page is still visible and online
            if (this.isPageVisible && navigator.onLine) {
                this.reconnectWebSocket(url, element);
            }
        }, delay);
    }

    reconnectWebSocket(url, element) {
        try {
            console.log(`[WebSocketManager] Attempting to reconnect ${url}`);

            // Trigger HTMX to reconnect by re-processing the hx-ws attribute
            if (element && element.getAttribute('hx-ws')) {
                // Remove and re-add the hx-ws attribute to trigger reconnection
                const wsAttribute = element.getAttribute('hx-ws');
                element.removeAttribute('hx-ws');

                // Small delay to ensure cleanup
                setTimeout(() => {
                    element.setAttribute('hx-ws', wsAttribute);
                    // Process the element with HTMX
                    if (window.htmx) {
                        window.htmx.process(element);
                    }
                }, 100);
            }
        } catch (error) {
            console.error('[WebSocketManager] Error during reconnection:', error);
        }
    }

    handleVisibilityChange() {
        const wasVisible = this.isPageVisible;
        this.isPageVisible = !document.hidden;

        if (!wasVisible && this.isPageVisible) {
            // Page became visible - check and reconnect websockets
            console.log('[WebSocketManager] Page became visible, checking websocket connections');
            this.handlePageVisible();
        } else if (wasVisible && !this.isPageVisible) {
            // Page became hidden
            console.log('[WebSocketManager] Page became hidden');
            this.handlePageHidden();
        }
    }

    handlePageVisible() {
        // Reset reconnect attempts when page becomes visible
        this.reconnectAttempts.clear();

        // Check all registered connections and reconnect if needed
        setTimeout(() => {
            this.checkAndReconnectAll();
        }, 500); // Small delay to ensure page is fully visible
    }

    handlePageHidden() {
        // Don't do anything special when page becomes hidden
        // Let the server handle the connection timeout gracefully
        console.log('[WebSocketManager] Page hidden, websocket connections will be maintained by server');
    }

    handleOnline() {
        console.log('[WebSocketManager] Browser came online');
        if (this.isPageVisible) {
            // Reset reconnect attempts and try to reconnect
            this.reconnectAttempts.clear();
            setTimeout(() => {
                this.checkAndReconnectAll();
            }, 1000);
        }
    }

    handleOffline() {
        console.log('[WebSocketManager] Browser went offline');
        // Reset reconnect attempts to avoid unnecessary reconnection attempts
        this.reconnectAttempts.clear();
    }

    checkAndReconnectAll() {
        if (!navigator.onLine || !this.isPageVisible) {
            return;
        }

        this.connections.forEach((connection, url) => {
            const element = connection.element;

            // Check if element still exists in DOM
            if (!document.contains(element)) {
                this.connections.delete(url);
                return;
            }

            // Try to reconnect if it's been a while since last connection
            const timeSinceLastConnection = Date.now() - connection.lastConnected;
            if (timeSinceLastConnection > 30000) { // 30 seconds
                console.log(`[WebSocketManager] Reconnecting stale connection: ${url}`);
                this.reconnectWebSocket(url, element);
            }
        });
    }

    // Public method to manually trigger reconnection of all websockets
    reconnectAll() {
        console.log('[WebSocketManager] Manually reconnecting all websockets');
        this.reconnectAttempts.clear();
        this.checkAndReconnectAll();
    }

    // Get connection statistics
    getStats() {
        return {
            totalConnections: this.connections.size,
            isPageVisible: this.isPageVisible,
            isOnline: navigator.onLine,
            reconnectAttempts: Object.fromEntries(this.reconnectAttempts)
        };
    }
}

// Initialize WebSocket Manager when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.webSocketManager = new WebSocketManager();

    // Add debug helper to window for testing
    window.wsManagerDebug = {
        stats: () => window.webSocketManager.getStats(),
        reconnectAll: () => window.webSocketManager.reconnectAll()
    };

    console.log('[WebSocketManager] WebSocket suspension handling initialized');
});

// Handle page focus/blur events as additional triggers
window.addEventListener('focus', () => {
    if (window.webSocketManager && navigator.onLine) {
        setTimeout(() => {
            window.webSocketManager.checkAndReconnectAll();
        }, 1000);
    }
});

// Periodic health check for websocket connections
setInterval(() => {
    if (window.webSocketManager && window.webSocketManager.isPageVisible && navigator.onLine) {
        window.webSocketManager.checkAndReconnectAll();
    }
}, 60000); // Check every minute
