/**
 * Image Preloader for Trouble Report Attachments
 *
 * This utility handles intelligent preloading of trouble report attachment images
 * to improve user experience by reducing loading times when users interact with images.
 *
 * Features:
 * - Automatic detection of visible images in viewport
 * - Preloading of images via service worker
 * - Intersection Observer for dynamic content
 * - Manual preloading API for specific images
 * - Debounced operations to avoid excessive requests
 */

class TroubleReportImagePreloader {
    constructor() {
        this.isServiceWorkerReady = false;
        this.preloadQueue = new Set();
        this.intersectionObserver = null;
        this.debounceTimer = null;
        this.initialized = false;

        // Configuration
        this.config = {
            // Preload images when they're within this distance from viewport
            rootMargin: '200px',
            // Minimum threshold for visibility
            threshold: 0.1,
            // Debounce delay for batch operations
            debounceDelay: 500,
            // Maximum images to preload per batch
            maxBatchSize: 10,
        };

        this.init();
    }

    /**
     * Initialize the preloader
     */
    async init() {
        if (this.initialized) return;

        console.log('[ImagePreloader] Initializing...');

        // Wait for service worker to be ready
        await this.waitForServiceWorker();

        // Set up intersection observer
        this.setupIntersectionObserver();

        // Preload currently visible images
        this.preloadVisibleImages();

        // Listen for dynamic content changes (HTMX updates)
        this.setupDynamicContentListeners();

        this.initialized = true;
        console.log('[ImagePreloader] Initialization complete');
    }

    /**
     * Wait for service worker to be available
     */
    async waitForServiceWorker() {
        if ('serviceWorker' in navigator) {
            try {
                const registration = await navigator.serviceWorker.ready;
                this.isServiceWorkerReady = true;
                console.log('[ImagePreloader] Service worker ready');
            } catch (error) {
                console.warn('[ImagePreloader] Service worker not available:', error);
            }
        } else {
            console.warn('[ImagePreloader] Service worker not supported');
        }
    }

    /**
     * Set up intersection observer to detect images entering viewport
     */
    setupIntersectionObserver() {
        if (!('IntersectionObserver' in window)) {
            console.warn('[ImagePreloader] IntersectionObserver not supported');
            return;
        }

        this.intersectionObserver = new IntersectionObserver(
            (entries) => {
                const imagesToPreload = [];

                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const imageUrl = this.extractImageUrl(entry.target);
                        if (imageUrl && !this.preloadQueue.has(imageUrl)) {
                            imagesToPreload.push(imageUrl);
                            this.preloadQueue.add(imageUrl);
                        }
                    }
                });

                if (imagesToPreload.length > 0) {
                    this.debouncedPreload(imagesToPreload);
                }
            },
            {
                rootMargin: this.config.rootMargin,
                threshold: this.config.threshold,
            }
        );

        // Observe existing images
        this.observeAttachmentImages();
    }

    /**
     * Set up listeners for dynamic content changes
     */
    setupDynamicContentListeners() {
        // Listen for HTMX content updates
        document.addEventListener('htmx:afterSwap', () => {
            console.log('[ImagePreloader] Content updated, re-observing images');
            this.observeAttachmentImages();
            this.preloadVisibleImages();
        });

        // Listen for new trouble report data loads
        document.addEventListener('htmx:afterRequest', (event) => {
            if (event.detail.requestConfig?.path?.includes('trouble-reports/data')) {
                console.log('[ImagePreloader] Trouble report data updated');
                setTimeout(() => {
                    this.observeAttachmentImages();
                    this.preloadVisibleImages();
                }, 100);
            }
        });
    }

    /**
     * Find and observe all attachment images in the document
     */
    observeAttachmentImages() {
        if (!this.intersectionObserver) return;

        // Find all attachment preview items
        const attachmentItems = document.querySelectorAll('.attachment-preview-item');

        attachmentItems.forEach(item => {
            const img = item.querySelector('.preview-thumbnail');
            if (img) {
                this.intersectionObserver.observe(img);
            }
        });

        console.log(`[ImagePreloader] Observing ${attachmentItems.length} attachment images`);
    }

    /**
     * Extract image URL from attachment element
     */
    extractImageUrl(imgElement) {
        const src = imgElement.src;
        if (src && src.includes('trouble-reports/attachments')) {
            return src;
        }

        // Fallback: construct URL from data attributes
        const attachmentItem = imgElement.closest('.attachment-preview-item');
        if (attachmentItem) {
            const reportId = attachmentItem.dataset.reportId;
            const attachmentId = attachmentItem.dataset.attachmentId;

            if (reportId && attachmentId) {
                return `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`;
            }
        }

        return null;
    }

    /**
     * Preload images that are currently visible
     */
    preloadVisibleImages() {
        const visibleImages = this.getVisibleAttachmentImages();
        if (visibleImages.length > 0) {
            console.log(`[ImagePreloader] Preloading ${visibleImages.length} visible images`);
            this.preloadImages(visibleImages);
        }
    }

    /**
     * Get images that are currently visible in the viewport
     */
    getVisibleAttachmentImages() {
        const images = [];
        const attachmentItems = document.querySelectorAll('.attachment-preview-item');

        attachmentItems.forEach(item => {
            if (this.isElementVisible(item)) {
                const imageUrl = this.extractImageUrl(item.querySelector('.preview-thumbnail'));
                if (imageUrl && !this.preloadQueue.has(imageUrl)) {
                    images.push(imageUrl);
                    this.preloadQueue.add(imageUrl);
                }
            }
        });

        return images;
    }

    /**
     * Check if element is visible in viewport
     */
    isElementVisible(element) {
        const rect = element.getBoundingClientRect();
        const windowHeight = window.innerHeight || document.documentElement.clientHeight;
        const windowWidth = window.innerWidth || document.documentElement.clientWidth;

        return (
            rect.top < windowHeight &&
            rect.bottom > 0 &&
            rect.left < windowWidth &&
            rect.right > 0
        );
    }

    /**
     * Debounced preload function to batch requests
     */
    debouncedPreload(imagesToAdd = []) {
        clearTimeout(this.debounceTimer);

        this.debounceTimer = setTimeout(() => {
            const allImages = Array.from(this.preloadQueue);

            // Limit batch size
            const batch = allImages.slice(0, this.config.maxBatchSize);

            if (batch.length > 0) {
                this.preloadImages(batch);
            }
        }, this.config.debounceDelay);
    }

    /**
     * Send images to service worker for preloading
     */
    async preloadImages(imageUrls) {
        if (!this.isServiceWorkerReady || !imageUrls.length) {
            return;
        }

        try {
            // Send message to service worker
            if (navigator.serviceWorker.controller) {
                navigator.serviceWorker.controller.postMessage({
                    type: 'PRELOAD_IMAGES',
                    images: imageUrls,
                });

                console.log(`[ImagePreloader] Requested preload of ${imageUrls.length} images`);
            }
        } catch (error) {
            console.error('[ImagePreloader] Failed to send preload request:', error);
        }
    }

    /**
     * Public API: Manually preload specific images
     */
    preloadSpecificImages(reportIds) {
        if (!Array.isArray(reportIds)) {
            reportIds = [reportIds];
        }

        const imagesToPreload = [];

        reportIds.forEach(reportId => {
            const attachmentItems = document.querySelectorAll(
                `.attachment-preview-item[data-report-id="${reportId}"]`
            );

            attachmentItems.forEach(item => {
                const imageUrl = this.extractImageUrl(item.querySelector('.preview-thumbnail'));
                if (imageUrl && !this.preloadQueue.has(imageUrl)) {
                    imagesToPreload.push(imageUrl);
                    this.preloadQueue.add(imageUrl);
                }
            });
        });

        if (imagesToPreload.length > 0) {
            console.log(`[ImagePreloader] Manual preload of ${imagesToPreload.length} images for reports:`, reportIds);
            this.preloadImages(imagesToPreload);
        }
    }

    /**
     * Public API: Clear preload queue
     */
    clearQueue() {
        this.preloadQueue.clear();
        console.log('[ImagePreloader] Preload queue cleared');
    }

    /**
     * Public API: Get current queue status
     */
    getQueueStatus() {
        return {
            size: this.preloadQueue.size,
            items: Array.from(this.preloadQueue),
        };
    }

    /**
     * Cleanup resources
     */
    destroy() {
        if (this.intersectionObserver) {
            this.intersectionObserver.disconnect();
            this.intersectionObserver = null;
        }

        clearTimeout(this.debounceTimer);
        this.preloadQueue.clear();
        this.initialized = false;

        console.log('[ImagePreloader] Destroyed');
    }
}

// Auto-initialize when DOM is ready
let imagePreloader = null;

function initImagePreloader() {
    if (!imagePreloader) {
        imagePreloader = new TroubleReportImagePreloader();
    }
}

// Initialize based on document state
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initImagePreloader);
} else {
    initImagePreloader();
}

// Export for manual usage
window.TroubleReportImagePreloader = TroubleReportImagePreloader;
window.imagePreloader = imagePreloader;

// Additional utility functions for integration with existing code
window.preloadTroubleReportImages = function(reportIds) {
    if (imagePreloader) {
        imagePreloader.preloadSpecificImages(reportIds);
    }
};

window.clearImagePreloadQueue = function() {
    if (imagePreloader) {
        imagePreloader.clearQueue();
    }
};
