class TroubleReportImagePreloader {
    constructor() {
        this.isServiceWorkerReady = false;
        this.preloadQueue = new Set();
        this.intersectionObserver = null;
        this.debounceTimer = null;
        this.initialized = false;
        this.config = {
            rootMargin: "200px",
            threshold: 0.1,
            debounceDelay: 500,
            maxBatchSize: 10,
        };
        this.init();
    }

    async init() {
        if (this.initialized) return;
        await this.waitForServiceWorker();
        this.setupIntersectionObserver();
        this.preloadVisibleImages();
        this.setupDynamicContentListeners();
        this.initialized = true;
    }

    async waitForServiceWorker() {
        if ("serviceWorker" in navigator) {
            try {
                await navigator.serviceWorker.ready;
                this.isServiceWorkerReady = true;
            } catch (error) {
                // Service worker not available
            }
        }
    }

    setupIntersectionObserver() {
        if (!("IntersectionObserver" in window)) return;

        this.intersectionObserver = new IntersectionObserver(
            (entries) => {
                const imagesToPreload = [];
                entries.forEach((entry) => {
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
            },
        );

        this.observeAttachmentImages();
    }

    setupDynamicContentListeners() {
        document.addEventListener("htmx:afterSwap", () => {
            this.observeAttachmentImages();
            this.preloadVisibleImages();
        });

        document.addEventListener("htmx:afterRequest", (event) => {
            if (
                event.detail.requestConfig?.path?.includes(
                    "trouble-reports/data",
                )
            ) {
                setTimeout(() => {
                    this.observeAttachmentImages();
                    this.preloadVisibleImages();
                }, 100);
            }
        });
    }

    observeAttachmentImages() {
        if (!this.intersectionObserver) return;
        const attachmentItems = document.querySelectorAll(
            ".attachment-preview-item",
        );
        attachmentItems.forEach((item) => {
            const img = item.querySelector(".preview-thumbnail");
            if (img) {
                this.intersectionObserver.observe(img);
            }
        });
    }

    extractImageUrl(imgElement) {
        const src = imgElement.src;
        if (src && src.includes("trouble-reports/attachments")) {
            return src;
        }

        const attachmentItem = imgElement.closest(".attachment-preview-item");
        if (attachmentItem) {
            const reportId = attachmentItem.dataset.reportId;
            const attachmentId = attachmentItem.dataset.attachmentId;
            if (reportId && attachmentId) {
                return `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`;
            }
        }
        return null;
    }

    preloadVisibleImages() {
        const visibleImages = this.getVisibleAttachmentImages();
        if (visibleImages.length > 0) {
            this.preloadImages(visibleImages);
        }
    }

    getVisibleAttachmentImages() {
        const images = [];
        const attachmentItems = document.querySelectorAll(
            ".attachment-preview-item",
        );

        attachmentItems.forEach((item) => {
            if (this.isElementVisible(item)) {
                const imageUrl = this.extractImageUrl(
                    item.querySelector(".preview-thumbnail"),
                );
                if (imageUrl && !this.preloadQueue.has(imageUrl)) {
                    images.push(imageUrl);
                    this.preloadQueue.add(imageUrl);
                }
            }
        });
        return images;
    }

    isElementVisible(element) {
        const rect = element.getBoundingClientRect();
        const windowHeight =
            window.innerHeight || document.documentElement.clientHeight;
        const windowWidth =
            window.innerWidth || document.documentElement.clientWidth;

        return (
            rect.top < windowHeight &&
            rect.bottom > 0 &&
            rect.left < windowWidth &&
            rect.right > 0
        );
    }

    debouncedPreload(imagesToAdd = []) {
        clearTimeout(this.debounceTimer);
        this.debounceTimer = setTimeout(() => {
            const allImages = Array.from(this.preloadQueue);
            const batch = allImages.slice(0, this.config.maxBatchSize);
            if (batch.length > 0) {
                this.preloadImages(batch);
            }
        }, this.config.debounceDelay);
    }

    async preloadImages(imageUrls) {
        if (!this.isServiceWorkerReady || !imageUrls.length) return;

        try {
            if (navigator.serviceWorker.controller) {
                navigator.serviceWorker.controller.postMessage({
                    type: "PRELOAD_IMAGES",
                    images: imageUrls,
                });
            }
        } catch (error) {
            console.error("Failed to send preload request:", error);
        }
    }

    preloadSpecificImages(reportIds) {
        if (!Array.isArray(reportIds)) {
            reportIds = [reportIds];
        }

        const imagesToPreload = [];
        reportIds.forEach((reportId) => {
            const attachmentItems = document.querySelectorAll(
                `.attachment-preview-item[data-report-id="${reportId}"]`,
            );

            attachmentItems.forEach((item) => {
                const imageUrl = this.extractImageUrl(
                    item.querySelector(".preview-thumbnail"),
                );
                if (imageUrl && !this.preloadQueue.has(imageUrl)) {
                    imagesToPreload.push(imageUrl);
                    this.preloadQueue.add(imageUrl);
                }
            });
        });

        if (imagesToPreload.length > 0) {
            this.preloadImages(imagesToPreload);
        }
    }

    clearQueue() {
        this.preloadQueue.clear();
    }

    getQueueStatus() {
        return {
            size: this.preloadQueue.size,
            items: Array.from(this.preloadQueue),
        };
    }

    destroy() {
        if (this.intersectionObserver) {
            this.intersectionObserver.disconnect();
            this.intersectionObserver = null;
        }
        clearTimeout(this.debounceTimer);
        this.preloadQueue.clear();
        this.initialized = false;
    }
}

let imagePreloader = null;

function initImagePreloader() {
    if (!imagePreloader) {
        imagePreloader = new TroubleReportImagePreloader();
    }
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initImagePreloader);
} else {
    initImagePreloader();
}

window.TroubleReportImagePreloader = TroubleReportImagePreloader;
window.imagePreloader = imagePreloader;

window.preloadTroubleReportImages = function (reportIds) {
    if (imagePreloader) {
        imagePreloader.preloadSpecificImages(reportIds);
    }
};

window.clearImagePreloadQueue = function () {
    if (imagePreloader) {
        imagePreloader.clearQueue();
    }
};
