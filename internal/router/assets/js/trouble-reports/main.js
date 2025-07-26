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

    // PDF Sharing functionality
    window.shareTroubleReportPDF = async function (troubleReportId, title) {
        let button = null;
        let originalContent = "";

        try {
            // Show loading state
            button = event.target.closest("button");
            if (!button) {
                alert("Fehler: Share-Button nicht gefunden.");
                return;
            }

            originalContent = button.innerHTML;
            button.innerHTML = '<i class="bi bi-hourglass-split"></i>';
            button.disabled = true;

            // Check HTTPS requirement for Web Share API
            const isHTTPS = location.protocol === "https:";
            const canUseWebShare = isHTTPS;

            console.log("=== PDF Share Decision Path ===");
            console.log("Current URL:", location.href);
            console.log("Protocol:", location.protocol);
            console.log("Hostname:", location.hostname);
            console.log("HTTPS:", isHTTPS);
            console.log("Can use Web Share API:", canUseWebShare);
            console.log("navigator.share exists:", "share" in navigator);
            console.log("navigator.canShare exists:", "canShare" in navigator);

            if (!canUseWebShare) {
                console.log(
                    "🚫 Decision: FORCE DIRECT DOWNLOAD - HTTPS required for Web Share API, skipping entirely",
                );
            } else {
                console.log(
                    "✅ Decision: HTTPS detected - will attempt Web Share API",
                );
            }

            // Fetch PDF from server
            const response = await fetch(
                `./trouble-reports/share-pdf?id=${troubleReportId}`,
            );

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const blob = await response.blob();
            const filename = `fehlerbericht_${troubleReportId}_${new Date().toISOString().split("T")[0]}.pdf`;

            // Decision: Use Web Share API only if HTTPS/localhost AND browser supports it
            if (canUseWebShare) {
                console.log(
                    "🔒 Secure context confirmed - checking Web Share API support...",
                );

                if (navigator.share && navigator.canShare) {
                    const file = new File([blob], filename, {
                        type: "application/pdf",
                    });
                    const shareData = {
                        title: `Fehlerbericht: ${title}`,
                        text: `Fehlerbericht #${troubleReportId}`,
                        files: [file],
                    };

                    // Check if we can share files and attempt to share
                    if (navigator.canShare(shareData)) {
                        console.log("Attempting Web Share API with file...");
                        try {
                            await navigator.share(shareData);
                            // Show success feedback
                            button.innerHTML =
                                '<i class="bi bi-check-circle"></i>';
                            button.style.color = "green";
                            setTimeout(() => {
                                button.innerHTML = originalContent;
                                button.style.color = "blue";
                                button.disabled = false;
                            }, 1500);
                            return; // Success - exit function
                        } catch (shareError) {
                            console.log("Web Share API failed:", shareError);

                            // Handle specific error types
                            if (shareError.name === "NotAllowedError") {
                                console.log("Share permission denied");
                            } else if (shareError.name === "AbortError") {
                                console.log("Share cancelled by user");
                            } else {
                                console.log(
                                    "Other share error:",
                                    shareError.name,
                                );
                            }

                            // Fall through to download
                        }
                    } else {
                        console.log(
                            "Cannot share files - falling back to download",
                        );
                    }
                } else {
                    console.log(
                        "Web Share API not available - falling back to download",
                    );
                }
            } else {
                console.log(
                    "🚫 Non-HTTPS connection - FORCING direct download, NO Web Share API attempt",
                );
                console.log(
                    "This prevents empty shares with missing PDF files",
                );
            }

            // Download the PDF (fallback or direct)
            console.log("📥 Starting PDF download...");
            console.log(
                "Reason:",
                canUseWebShare
                    ? "Web Share API failed/unsupported"
                    : "Non-HTTPS connection - direct download",
            );
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = filename;
            a.style.display = "none";
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            // Show download success feedback
            button.innerHTML = '<i class="bi bi-download"></i>';
            button.style.color = "green";
            setTimeout(() => {
                button.innerHTML = originalContent;
                button.style.color = "blue";
                button.disabled = false;
            }, 1500);
        } catch (error) {
            console.error("Error sharing/downloading PDF:", error);
            alert(
                "Fehler beim Erstellen oder Teilen der PDF. Bitte versuchen Sie es erneut.",
            );
            // Restore button state on error
            if (button && originalContent) {
                button.innerHTML = originalContent;
                button.style.color = "blue";
                button.disabled = false;
            }
        }
    };
})();
