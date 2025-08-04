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
