(() => {
    // Update share button appearance based on HTTPS/HTTP
    const isHTTPS = location.protocol === "https:";

    document.querySelectorAll('[id^="share-btn-"]').forEach((button) => {
        if (!isHTTPS) {
            button.title = "PDF herunterladen (HTTPS erforderlich für Teilen)";
            const icon = button.querySelector("i");
            if (icon) {
                icon.className = "bi bi-download";
            }
        } else {
            button.title = "Als PDF teilen";
        }
    });

    function loadAttachmentsForTroubleReport(troubleReportId) {
        const placeholder = document.querySelector(
            `[data-trouble-report-id="${troubleReportId}"]`,
        );
        if (placeholder) {
            htmx.trigger(document.body, `load-attachments-${troubleReportId}`);
        }
    }

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
    handleHashNavigation();

    // Add event listeners for details toggle
    document
        .querySelectorAll('details[id^="trouble-report-"]')
        .forEach((details) => {
            details.addEventListener("toggle", function () {
                if (this.open) {
                    // Close all other details tags
                    document
                        .querySelectorAll('details[id^="trouble-report-"]')
                        .forEach((otherDetails) => {
                            if (otherDetails !== this && otherDetails.open) {
                                otherDetails.open = false;
                            }
                        });

                    // Update location hash and load attachments
                    history.replaceState(null, null, `#${this.id}`);
                    const troubleReportId = this.id.replace(
                        "trouble-report-",
                        "",
                    );
                    loadAttachmentsForTroubleReport(troubleReportId);
                } else {
                    history.replaceState(
                        null,
                        null,
                        location.pathname + location.search,
                    );
                }
            });
        });
})();
