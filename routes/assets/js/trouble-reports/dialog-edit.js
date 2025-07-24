document.querySelector("#dialogEdit").showModal();

// Initialize variables and functions
(function () {
    const html = String.raw;

    // Local variables to avoid global conflicts
    let attachmentOrder = [];
    let selectedFiles = [];

    // Reset file state to prevent duplicates on dialog reopen
    function resetFileState() {
        selectedFiles = [];
        const fileInput = document.getElementById("attachments");
        if (fileInput) {
            fileInput.value = "";
        }
        const previewArea = document.getElementById("file-preview");
        if (previewArea) {
            previewArea.style.display = "none";
        }
        const container = document.getElementById("new-attachments");
        if (container) {
            container.innerHTML = "";
        }
    }

    // Check if there are validation errors in the form
    function hasValidationErrors() {
        const titleInput = document.getElementById("title");
        const contentInput = document.getElementById("content");

        return (
            (titleInput &&
                titleInput.getAttribute("aria-invalid") === "true") ||
            (contentInput &&
                contentInput.getAttribute("aria-invalid") === "true")
        );
    }

    // Update hidden input with current attachment order
    function updateAttachmentOrderInput() {
        document.getElementById("attachment-order").value =
            attachmentOrder.join(",");
    }

    // Initialize attachment order from existing attachments
    function initializeAttachmentOrder() {
        // Initialize sortable for existing attachments
        if (
            window.Sortable &&
            document.getElementById("existing-attachments")
        ) {
            new Sortable(document.getElementById("existing-attachments"), {
                animation: 150,
                ghostClass: "sortable-ghost",
                handle: ".bi-grip-vertical",
                onEnd: function (evt) {
                    // Update attachment order
                    const items = document.querySelectorAll(
                        "#existing-attachments .attachment-item",
                    );
                    attachmentOrder = Array.from(items).map(
                        (item) => item.dataset.id,
                    );
                    updateAttachmentOrderInput();
                },
            });
        }

        const existingAttachments = document.querySelectorAll(
            "#existing-attachments .attachment-item",
        );
        attachmentOrder = Array.from(existingAttachments).map(
            (item) => item.dataset.id,
        );
        updateAttachmentOrderInput();
    }

    // Update the file input with current selected files
    function updateFileInput() {
        const fileInput = document.getElementById("attachments");
        const dt = new DataTransfer();

        selectedFiles.forEach((file) => {
            dt.items.add(file);
        });

        fileInput.files = dt.files;
    }

    // Display file preview
    function displayFilePreview() {
        const previewArea = document.getElementById("file-preview");
        const container = document.getElementById("new-attachments");
        container.innerHTML = "";

        if (selectedFiles.length > 0) {
            previewArea.style.display = "block";

            selectedFiles.forEach((file, index) => {
                const sizeClass =
                    file.size > 10 * 1024 * 1024
                        ? "attachment-error"
                        : "text-muted";
                const sizeText =
                    file.size > 10 * 1024 * 1024
                        ? "ZU GROSS!"
                        : formatFileSize(file.size);

                const item = document.createElement("div");
                item.className = "attachment-item";
                item.innerHTML = html`
                    <div class="attachment-info">
                        <i class="bi bi-file-earmark attachment-icon"></i>

                        <span class="ellipsis">${file.name}</span>

                        <span class="${sizeClass}">(${sizeText})</span>
                    </div>

                    <div class="attachment-actions">
                        <button
                            type="button"
                            class="destructive flex gap"
                            onclick="window.dialogEditFunctions.removeFileFromPreview(${index})"
                        >
                            <small>
                                <i class="bi bi-trash"></i>
                                Entfernen
                            </small>
                        </button>
                    </div>
                `;
                container.appendChild(item);
            });
        } else {
            previewArea.style.display = "none";
        }
    }

    // Format file size
    function formatFileSize(bytes) {
        if (bytes === 0) return "0 Bytes";
        const k = 1024;
        const sizes = ["Bytes", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }

    // Expose functions globally that need to be called from HTML
    window.dialogEditFunctions = {
        handleFileSelect: function (event) {
            selectedFiles = Array.from(event.target.files);
            displayFilePreview();
        },

        handleFileDrop: function (event) {
            event.preventDefault();
            event.stopPropagation();
            const area = event.currentTarget;
            area.classList.remove("dragover");

            selectedFiles = Array.from(event.dataTransfer.files);
            updateFileInput();
            displayFilePreview();
        },

        handleDragOver: function (event) {
            event.preventDefault();
            event.currentTarget.classList.add("dragover");
        },

        handleDragLeave: function (event) {
            event.currentTarget.classList.remove("dragover");
        },

        removeFileFromPreview: function (index) {
            selectedFiles.splice(index, 1);
            updateFileInput();
            displayFilePreview();
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

        showImageDialog: function (reportId, attachmentId) {
            const imageUrl = `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`;

            // Create image dialog
            const dialog = document.createElement("dialog");
            dialog.className = "image-viewer-dialog no-user-select";
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
                    <div class="image-viewer-body no-user-select">
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

        deleteAttachment: function (reportId, attachmentId) {
            if (
                confirm(
                    "Sind Sie sicher, dass Sie diesen Anhang löschen möchten?",
                )
            ) {
                htmx.ajax(
                    "DELETE",
                    `./trouble-reports/attachments?id=${reportId}&attachment_id=${attachmentId}`,
                    {
                        target: "#attachments-section",
                        swap: "innerHTML",
                    },
                );
            }
        },
    };

    // Reset file state and initialize attachment order on load
    // If there are validation errors, always reset to clear any uploaded files
    if (hasValidationErrors()) {
        resetFileState();
    } else {
        resetFileState();
    }
    initializeAttachmentOrder();

    // Form validation
    document.querySelector("form").addEventListener("submit", function (e) {
        let hasOversizedFiles = false;

        selectedFiles.forEach((file) => {
            if (file.size > 10 * 1024 * 1024) {
                hasOversizedFiles = true;
            }
        });

        if (hasOversizedFiles) {
            e.preventDefault();
            alert(
                "Einige Dateien sind zu groß. Die maximale Dateigröße beträgt 10MB.",
            );
            return false;
        }

        const totalFiles =
            selectedFiles.length +
            (document.querySelectorAll("#existing-attachments .attachment-item")
                .length || 0);
        if (totalFiles > 10) {
            e.preventDefault();
            alert("Zu viele Anhänge. Maximal 10 Anhänge sind erlaubt.");
            return false;
        }
    });
})();

// Global functions for backward compatibility
function handleFileSelect(event) {
    window.dialogEditFunctions.handleFileSelect(event);
}
function handleFileDrop(event) {
    window.dialogEditFunctions.handleFileDrop(event);
}
function handleDragOver(event) {
    window.dialogEditFunctions.handleDragOver(event);
}
function handleDragLeave(event) {
    window.dialogEditFunctions.handleDragLeave(event);
}
function viewAttachment(reportId, attachmentId, isImage) {
    window.dialogEditFunctions.viewAttachment(reportId, attachmentId, isImage);
}
function deleteAttachment(reportId, attachmentId) {
    window.dialogEditFunctions.deleteAttachment(reportId, attachmentId);
}
