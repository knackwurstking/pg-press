(() => {
    function loadAttachmentsForModification(troubleReportId, modificationTime) {
        const placeholder = document.querySelector(
            `[data-modification-id="${troubleReportId}-${modificationTime}"]`,
        );
        if (placeholder) {
            htmx.trigger(
                document.body,
                `load-mod-attachments-${troubleReportId}-${modificationTime}`,
            );
        }
    }

    document
        .querySelectorAll(
            ".attachments-preview-placeholder[data-modification-id]",
        )
        .forEach((placeholder) => {
            placeholder.style.cursor = "pointer";
            placeholder.style.transition = "all 0.2s ease";
            placeholder.style.borderStyle = "solid";

            placeholder.addEventListener("click", function () {
                const modificationId = this.getAttribute(
                    "data-modification-id",
                );
                const [troubleReportId, modificationTime] =
                    modificationId.split("-");
                loadAttachmentsForModification(
                    troubleReportId,
                    modificationTime,
                );
            });

            placeholder.addEventListener("mouseenter", function () {
                this.style.backgroundColor = "var(--ui-info)";
                this.style.borderColor = "var(--ui-primary)";
                this.style.color = "var(--ui-info-text)";
            });

            placeholder.addEventListener("mouseleave", function () {
                this.style.backgroundColor = "";
                this.style.borderColor = "var(--ui-border-color)";
                this.style.color = "";
            });
        });
})();
