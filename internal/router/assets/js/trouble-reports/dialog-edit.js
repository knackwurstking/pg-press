document.querySelector("#dialogEdit").showModal();

(() => {
    let attachmentOrder = [];
    let selectedFiles = [];

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

    function initializeAttachmentOrder() {
        if (
            window.Sortable &&
            document.getElementById("existing-attachments")
        ) {
            new Sortable(document.getElementById("existing-attachments"), {
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
                        : "text-muted";
                const sizeText =
                    file.size > 10 * 1024 * 1024
                        ? "ZU GROSS!"
                        : formatFileSize(file.size);

                const item = document.createElement("div");
                item.className = "attachment-item";
                item.innerHTML = `
                    <div class="attachment-info">
                        <i class="bi bi-file-earmark attachment-icon"></i>
                        <span class="ellipsis">${file.name}</span>
                        <span class="${sizeClass}">(${sizeText})</span>
                    </div>
                    <div class="attachment-actions">
                        <button type="button" class="destructive flex gap" onclick="window.dialogEditFunctions.removeFileFromPreview(${index})">
                            <small><i class="bi bi-trash"></i> Entfernen</small>
                        </button>
                    </div>
                `;
                container.appendChild(item);
            });
        } else {
            previewArea.style.display = "none";
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
            event.currentTarget.classList.remove("dragover");
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
            window.TroubleReportsImageViewer.viewAttachment(
                reportId,
                attachmentId,
                isImage,
            );
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
})();
