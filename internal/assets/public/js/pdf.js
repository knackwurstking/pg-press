async function downloadPDF(event, url, fallbackFilename) {
	var button = event.target.closest('button') || event.target;

	try {
		// Disable button to prevent multiple clicks
		button.disabled = true;

		// Fetch the PDF from the server
		var response = await fetch(url);
		if (!response.ok) {
			throw new Error('PDF konnte nicht geladen werden');
		}

		// Get the blob and create download
		var blob = await response.blob();
		var downloadUrl = window.URL.createObjectURL(blob);
		var a = document.createElement('a');

		// Extract filename from headers or use the fallback name
		var contentDisposition = response.headers.get('Content-Disposition');
		var filenameMatch = contentDisposition?.match(/filename="(.+)"/);
		var filename = filenameMatch?.[1] || fallbackFilename;

		// Configure and trigger download
		a.style.display = "none"
		a.href = downloadUrl;
		a.download = filename;

		document.body.appendChild(a);
		a.click();

		// Cleanup
		window.URL.revokeObjectURL(downloadUrl);
		document.body.removeChild(a);
	} catch (error) {
		console.error('Download failed:', error);
		alert('Fehler beim Download: ' + error.message);
	} finally {
		// Ensure button is re-enabled in case of any unexpected errors
		button.disabled = false;
	}
}
