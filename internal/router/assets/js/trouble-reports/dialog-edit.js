document.querySelector("#dialogEdit").showModal();

(() => {
    let attachmentOrder = [];
    let selectedFiles = [];
    let isDragging = false;

    // File state management
    function resetFileState() {
        selectedFiles = [];
        const fileInput = document.getElementById("attachments");
        if (fileInput) fileInput.value = "";

        const previewArea = document.getElementById("file-preview");
        if (previewArea) previewArea.style.display = "none";

        const container = document.getElementById("new-attachments");
        if (container) container.innerHTML = "";
    }

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

    function updateAttachmentOrderInput() {
        document.getElementById("attachment-order").value =
            attachmentOrder.join(",");
    }

    let sortableInstance = null;

    function initializeAttachmentOrder() {
        const existingAttachmentsContainer = document.getElementById(
            "existing-attachments",
        );

        // Destroy existing Sortable instance if it exists
        if (sortableInstance) {
            sortableInstance.destroy();
            sortableInstance = null;
        }

        if (window.Sortable && existingAttachmentsContainer) {
            sortableInstance = new Sortable(existingAttachmentsContainer, {
                animation: 150,
                ghostClass: "sortable-ghost",
                handle: ".bi-grip-vertical",
                onEnd: function () {
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

    function updateFileInput() {
        const fileInput = document.getElementById("attachments");
        const dt = new DataTransfer();
        selectedFiles.forEach((file) => dt.items.add(file));
        fileInput.files = dt.files;
    }

    function formatFileSize(bytes) {
        if (bytes === 0) return "0 Bytes";
        const k = 1024;
        const sizes = ["Bytes", "KB", "MB", "GB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }

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
                        : "muted text-sm";
                const sizeText =
                    file.size > 10 * 1024 * 1024
                        ? "ZU GROSS!"
                        : formatFileSize(file.size);

                const item = document.createElement("div");
                item.className =
                    "attachment-item flex row gap justify-between align-center";
                item.innerHTML = `
                    <div class="attachment-info flex row gap align-center">
                        <i class="bi bi-file-earmark attachment-icon"></i>
                        <span class="ellipsis">${file.name}</span>
                        <span class="${sizeClass}">(${sizeText})</span>
                    </div>
                    <div class="attachment-actions flex row gap">
                        <button type="button" class="destructive flex row gap align-center" onclick="window.dialogEditFunctions.removeFileFromPreview(${index})">
                            <small class="flex row gap align-center"><i class="bi bi-trash"></i> Entfernen</small>
                        </button>
                    </div>
                `;
                container.appendChild(item);
            });
        } else {
            previewArea.style.display = "none";
        }
    }

    // Scroll prevention functions using ui.min.css classes
    function disableScroll() {
        document.body.classList.add("no-scrollbar", "no-user-select");
        const dialog = document.querySelector("#dialogEdit");
        if (dialog) {
            dialog.classList.add("no-scrollbar", "no-user-select");
        }
    }

    function enableScroll() {
        document.body.classList.remove("no-scrollbar", "no-user-select");
        const dialog = document.querySelector("#dialogEdit");
        if (dialog) {
            dialog.classList.remove("no-scrollbar", "no-user-select");
        }
    }

    // Global functions for HTML event handlers
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
            isDragging = false;
            enableScroll();
            selectedFiles = Array.from(event.dataTransfer.files);
            updateFileInput();
            displayFilePreview();
        },

        handleDragOver: function (event) {
            event.preventDefault();
            event.stopPropagation();
            const area = event.currentTarget;
            area.classList.add("dragover");
            if (!isDragging) {
                isDragging = true;
                disableScroll();
            }
        },

        handleDragLeave: function (event) {
            event.preventDefault();
            event.stopPropagation();
            const area = event.currentTarget;

            // Check if we're actually leaving the drag area
            const rect = area.getBoundingClientRect();
            const x = event.clientX;
            const y = event.clientY;

            if (
                x < rect.left ||
                x > rect.right ||
                y < rect.top ||
                y > rect.bottom
            ) {
                area.classList.remove("dragover");
                isDragging = false;
                enableScroll();
            }
        },

        removeFileFromPreview: function (index) {
            selectedFiles.splice(index, 1);
            updateFileInput();
            displayFilePreview();
        },

        viewAttachment: function (reportId, attachmentId, isImage) {
            window.TroubleReportsImageViewer.viewAttachment(
                reportId,
                attachmentId,
                isImage,
            );
        },

        deleteAttachment: function (attachmentId) {
            if (
                confirm(
                    "Sind Sie sicher, dass Sie diesen Anhang löschen möchten?",
                )
            ) {
                // Find and remove the attachment item from DOM
                const attachmentItem = document.querySelector(
                    `#existing-attachments .attachment-item[data-id="${attachmentId}"]`,
                );

                if (attachmentItem) {
                    // Remove from attachmentOrder array
                    attachmentOrder = attachmentOrder.filter(
                        (id) => id != attachmentId,
                    );

                    // Temporarily disable htmx processing during DOM manipulation
                    if (window.htmx) {
                        const container = document.getElementById(
                            "existing-attachments",
                        );
                        const wasDisabled =
                            container.hasAttribute("data-hx-disable");
                        container.setAttribute("data-hx-disable", "true");

                        // Remove the DOM element
                        attachmentItem.remove();

                        // Restore htmx state
                        if (!wasDisabled) {
                            container.removeAttribute("data-hx-disable");
                        }
                    } else {
                        // Remove the DOM element
                        attachmentItem.remove();
                    }

                    // Update the hidden input field
                    updateAttachmentOrderInput();

                    // Reinitialize Sortable instance after DOM changes
                    initializeAttachmentOrder();

                    // Check if no attachments left and hide the details section
                    const existingAttachments = document.getElementById(
                        "existing-attachments",
                    );
                    if (
                        existingAttachments &&
                        existingAttachments.children.length === 0
                    ) {
                        const detailsSection =
                            existingAttachments.closest("details");
                        if (detailsSection) {
                            detailsSection.style.display = "none";
                        }
                    }
                }
            }
        },
    };

    // Initialize
    if (hasValidationErrors()) {
        resetFileState();
    } else {
        resetFileState();
    }
    initializeAttachmentOrder();

    // Form validation
    document.querySelector("form").addEventListener("submit", function (e) {
        let hasOversizedFiles = selectedFiles.some(
            (file) => file.size > 10 * 1024 * 1024,
        );

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

    // Cleanup scroll prevention on dialog close
    const dialog = document.querySelector("#dialogEdit");
    if (dialog) {
        dialog.addEventListener("close", function () {
            if (isDragging) {
                isDragging = false;
                enableScroll();
            }
        });
    }

    // Add global drag end listener as fallback
    document.addEventListener("dragend", function () {
        if (isDragging) {
            isDragging = false;
            enableScroll();
        }
    });
})();
