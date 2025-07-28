// Shared Trouble Reports Functionality
window.TroubleReportsImageViewer = {
    showImageDialog: function (reportId, attachmentId) {
        const imageUrl = `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`;
        const dialog = document.createElement("dialog");
        dialog.className = "image-viewer-dialog";

        dialog.innerHTML = `
            <div class="image-viewer-content">
                <div class="image-viewer-header">
                    <button type="button" class="close-image-viewer" onclick="this.closest('dialog').close()">
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
        });

        img.addEventListener("error", () => {
            loadingDiv.innerHTML =
                '<i class="bi bi-exclamation-triangle"></i><span>Fehler beim Laden des Bildes</span>';
        });

        img.addEventListener("click", (e) => {
            e.stopPropagation();
            img.classList.toggle("zoomed");
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
            if (e.key === "Escape") dialog.close();
        });

        dialog.addEventListener("close", () =>
            document.body.removeChild(dialog),
        );

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
