(() => {
    // Update share button appearance based on HTTPS/HTTP
    const isHTTPS = location.protocol === "https:";
    const canUseWebShare = isHTTPS;

    document.querySelectorAll('[id^="share-btn-"]').forEach((button) => {
        if (!canUseWebShare) {
            // HTTP connection - change to download appearance
            button.title = "PDF herunterladen (HTTPS erforderlich für Teilen)";
            const icon = button.querySelector("i");
            if (icon) {
                icon.className = "bi bi-download";
            }
        } else {
            // HTTPS connection - keep share appearance
            button.title = "Als PDF teilen";
        }
    });

    // Function to handle attachment loading
    function loadAttachmentsForTroubleReport(troubleReportId) {
        const placeholder = document.querySelector(
            `[data-trouble-report-id="${troubleReportId}"]`,
        );
        if (placeholder) {
            htmx.trigger(document.body, `load-attachments-${troubleReportId}`);
        }
    }

    // Handle hash-based navigation after HTMX has loaded
    function handleHashNavigation() {
        if (location.hash) {
            const el = document.querySelector(location.hash);
            if (el) {
                el.open = true;
                const troubleReportId = el.id.replace("trouble-report-", "");
                loadAttachmentsForTroubleReport(troubleReportId);
            }
        }
    }

    // Run hash navigation after HTMX has finished processing
    document.addEventListener("htmx:load", handleHashNavigation);

    // Also run immediately in case HTMX has already loaded
    handleHashNavigation();

    // Add event listeners for details toggle
    document
        .querySelectorAll('details[id^="trouble-report-"]')
        .forEach((details) => {
            details.addEventListener("toggle", function () {
                if (this.open) {
                    const troubleReportId = this.id.replace(
                        "trouble-report-",
                        "",
                    );
                    loadAttachmentsForTroubleReport(troubleReportId);
                }
            });
        });

    // PDF sharing functionality is now in main.js
})();
