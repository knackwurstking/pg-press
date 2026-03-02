function openImageViewer(imageURL) {
	// Close any existing image viewer
	var existingDialog = document.querySelector('dialog[name="image-viewer"]');
	if (existingDialog) {
		existingDialog.close();
	}

	// Create new dialog from template
	var template = document.querySelector('template[name="image-viewer"]');
	var dialog = template.content.cloneNode(true).querySelector('dialog');
	var img = dialog.querySelector('img.attachment');

	// Setup image
	img.src = imageURL;
	img.onclick = function(event) {
		event.stopPropagation();
	};

	// Setup dialog click handler
	dialog.onclick = function() {
		dialog.close();
	};

	// Cleanup on close
	dialog.onclose = function() {
		document.body.removeChild(dialog);
	};

	// Show dialog
	document.body.appendChild(dialog);
	dialog.showModal();
}
