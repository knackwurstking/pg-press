// Shared Trouble Reports Functionality
window.TroubleReportsImageViewer = {
    showImageDialog: function (reportId, attachmentId) {
        const imageUrl = `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`;
        const dialog = document.createElement("dialog");
        dialog.className = "image-viewer-dialog";

        dialog.innerHTML = `
            <div class="image-viewer-content no-user-select">
                <div class="image-viewer-header">
                    <button
                        type="button"
                        class="icon secondary ghost"
                        onclick="this.closest('dialog').close()"
                        title="Schließen"
                    >
                        <i class="bi bi-x-lg"></i>
                    </button>
                </div>
                <div class="image-viewer-body">
                    <div class="image-loading">
                        <i class="bi bi-hourglass-split"></i>
                        <span>Bild wird geladen...</span>
                    </div>
                    <img src="${imageUrl}" alt="Attachment" class="fullscreen-image" style="display: none;" />
                </div>
            </div>
        `;

        const img = dialog.querySelector(".fullscreen-image");
        const loadingDiv = dialog.querySelector(".image-loading");

        img.addEventListener("load", () => {
            loadingDiv.style.display = "none";
            img.style.display = "block";
            // Ensure image fits viewport properly
            fitImageToViewport(img);
        });

        img.addEventListener("error", () => {
            loadingDiv.innerHTML =
                '<i class="bi bi-exclamation-triangle"></i><span>Fehler beim Laden des Bildes</span>';
        });

        img.addEventListener("click", (e) => {
            e.stopPropagation();
            toggleImageZoom(img);
        });

        // Double-click for zoom toggle
        img.addEventListener("dblclick", (e) => {
            e.stopPropagation();
            toggleImageZoom(img);
        });

        dialog.addEventListener("click", (e) => {
            if (
                e.target === dialog ||
                e.target.classList.contains("image-viewer-body")
            ) {
                dialog.close();
            }
        });

        dialog.addEventListener("keydown", (e) => {
            switch (e.key) {
                case "Escape":
                    dialog.close();
                    break;
                case " ":
                case "Enter":
                    e.preventDefault();
                    toggleImageZoom(img);
                    break;
                case "0":
                    e.preventDefault();
                    resetImageZoom(img);
                    break;
                case "+":
                case "=":
                    e.preventDefault();
                    zoomImageIn(img);
                    break;
                case "-":
                    e.preventDefault();
                    zoomImageOut(img);
                    break;
            }
        });

        // Prevent context menu on right-click
        dialog.addEventListener("contextmenu", (e) => {
            e.preventDefault();
            e.stopPropagation();
            return false;
        });

        img.addEventListener("contextmenu", (e) => {
            e.preventDefault();
            e.stopPropagation();
            return false;
        });

        dialog.addEventListener("close", () =>
            document.body.removeChild(dialog),
        );

        // Add no-user-select class to dialog
        dialog.classList.add("no-user-select");

        document.body.appendChild(dialog);
        dialog.showModal();
    },

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

// PDF Sharing functionality
window.shareTroubleReportPDF = async function (troubleReportId, title) {
    let button = null;
    let originalContent = "";

    try {
        button = event.target.closest("button");
        if (!button) {
            alert("Fehler: Share-Button nicht gefunden.");
            return;
        }

        originalContent = button.innerHTML;
        button.innerHTML = '<i class="bi bi-hourglass-split"></i>';
        button.disabled = true;

        const isHTTPS = location.protocol === "https:";
        const response = await fetch(
            `./trouble-reports/share-pdf?id=${troubleReportId}`,
        );

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const blob = await response.blob();
        const filename = `fehlerbericht_${troubleReportId}_${new Date().toISOString().split("T")[0]}.pdf`;

        // Try Web Share API first (only on HTTPS)
        if (isHTTPS && navigator.share && navigator.canShare) {
            const file = new File([blob], filename, {
                type: "application/pdf",
            });
            const shareData = {
                title: `Fehlerbericht: ${title}`,
                text: `Fehlerbericht #${troubleReportId}`,
                files: [file],
            };

            if (navigator.canShare(shareData)) {
                try {
                    await navigator.share(shareData);
                    button.innerHTML = '<i class="bi bi-check-circle"></i>';
                    button.style.color = "green";
                    setTimeout(() => {
                        button.innerHTML = originalContent;
                        button.style.color = "blue";
                        button.disabled = false;
                    }, 1500);
                    return;
                } catch (shareError) {
                    // Fall through to download
                }
            }
        }

        // Download fallback
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        a.style.display = "none";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

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
        if (button && originalContent) {
            button.innerHTML = originalContent;
            button.style.color = "blue";
            button.disabled = false;
        }
    }
};

// Image viewer utility functions
function fitImageToViewport(img) {
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const closeButtonSpace = 60; // Space for close button and padding

    // Calculate available space
    const maxWidth = viewportWidth - 40; // 20px padding on each side
    const maxHeight = viewportHeight - closeButtonSpace;

    // Get actual image dimensions
    const naturalWidth = img.naturalWidth;
    const naturalHeight = img.naturalHeight;

    if (naturalWidth && naturalHeight) {
        // Calculate scale to fit
        const scaleX = maxWidth / naturalWidth;
        const scaleY = maxHeight / naturalHeight;
        const scale = Math.min(scaleX, scaleY, 1); // Don't scale up beyond natural size

        // Apply calculated dimensions
        img.style.width = Math.floor(naturalWidth * scale) + "px";
        img.style.height = Math.floor(naturalHeight * scale) + "px";
    }
}

function toggleImageZoom(img) {
    if (img.classList.contains("zoomed")) {
        resetImageZoom(img);
    } else {
        zoomImageIn(img);
    }
}

function resetImageZoom(img) {
    img.classList.remove("zoomed");
    img.style.transform = "scale(1)";
    fitImageToViewport(img);
}

function zoomImageIn(img) {
    img.classList.add("zoomed");
    const currentScale = getImageScale(img);
    const newScale = Math.min(currentScale * 1.5, 3); // Max 3x zoom
    img.style.transform = `scale(${newScale})`;
}

function zoomImageOut(img) {
    const currentScale = getImageScale(img);
    const newScale = Math.max(currentScale / 1.5, 0.5); // Min 0.5x zoom
    img.style.transform = `scale(${newScale})`;

    if (newScale <= 1) {
        img.classList.remove("zoomed");
        fitImageToViewport(img);
    }
}

function getImageScale(img) {
    const matrix = window.getComputedStyle(img).transform;
    if (matrix === "none" || matrix === undefined) {
        return 1;
    }

    const matrixValues = matrix.match(/matrix.*\((.+)\)/)[1].split(", ");
    return parseFloat(matrixValues[0]) || 1;
}

// Global functions for backward compatibility
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
