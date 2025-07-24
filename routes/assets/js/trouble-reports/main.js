// Shared Trouble Reports Functionality
// This file contains common functionality used across trouble reports pages

(function () {
    const html = String.raw;

    // Shared image viewer dialog functionality
    window.TroubleReportsImageViewer = {
        showImageDialog: function (reportId, attachmentId) {
            const imageUrl = `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`;

            // Create image dialog
            const dialog = document.createElement("dialog");
            dialog.className = "image-viewer-dialog";
            dialog.innerHTML = html`
                <div class="image-viewer-content">
                    <div class="image-viewer-header">
                        <button
                            type="button"
                            class="close-image-viewer"
                            onclick="this.closest('dialog').close()"
                        >
                            <i class="bi bi-x-lg"></i>
                        </button>
                    </div>
                    <div class="image-viewer-body">
                        <div class="image-loading">
                            <i class="bi bi-hourglass-split"></i>
                            <span>Bild wird geladen...</span>
                        </div>
                        <img
                            src="${imageUrl}"
                            alt="Attachment"
                            class="fullscreen-image"
                            style="display: none;"
                        />
                    </div>
                </div>
            `;

            // Get image element and add load/error handlers
            const img = dialog.querySelector(".fullscreen-image");
            const loadingDiv = dialog.querySelector(".image-loading");

            img.addEventListener("load", function () {
                loadingDiv.style.display = "none";
                img.style.display = "block";
            });

            img.addEventListener("error", function () {
                loadingDiv.innerHTML =
                    '<i class="bi bi-exclamation-triangle"></i><span>Fehler beim Laden des Bildes</span>';
            });

            // Add click-to-zoom functionality
            img.addEventListener("click", function (e) {
                e.stopPropagation();
                img.classList.toggle("zoomed");
            });

            // Add click outside to close
            dialog.addEventListener("click", function (e) {
                if (
                    e.target === dialog ||
                    e.target.classList.contains("image-viewer-body")
                ) {
                    dialog.close();
                }
            });

            // Add escape key handler
            dialog.addEventListener("keydown", function (e) {
                if (e.key === "Escape") {
                    dialog.close();
                }
            });

            // Clean up on close
            dialog.addEventListener("close", function () {
                document.body.removeChild(dialog);
            });

            document.body.appendChild(dialog);
            dialog.showModal();
        },

        // Shared attachment viewing logic
        viewAttachment: function (reportId, attachmentId, isImage) {
            if (isImage === "true") {
                this.showImageDialog(reportId, attachmentId);
            } else {
                window.open(
                    `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`,
                    "_blank",
                );
            }
        },
    };

    // Global functions for backward compatibility and easy access
    window.viewAttachmentFromData = function (reportId, attachmentId, isImage) {
        window.TroubleReportsImageViewer.viewAttachment(
            reportId,
            attachmentId,
            isImage,
        );
    };

    window.viewAttachment = function (reportId, attachmentId, isImage) {
        window.TroubleReportsImageViewer.viewAttachment(
            reportId,
            attachmentId,
            isImage,
        );
    };
})();
