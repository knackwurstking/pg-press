var selectedFiles = [];
var existingAttachmentsRemoval = [];
var isPreviewFullscreen = false;

document.addEventListener('DOMContentLoaded', initializeMarkdownFeatures);

function initializeMarkdownFeatures() {
	var checkbox = document.getElementById('use_markdown');
	if (checkbox) {
		toggleMarkdownFeatures();
	}
}

function toggleMarkdownFeatures() {
	var checkbox = document.getElementById('use_markdown');
	var previewContainer = document.getElementById('markdown-preview-container');
	var textarea = document.getElementById('content');

	if (checkbox.checked) {
		previewContainer.classList.remove("hidden")
		textarea.setAttribute('placeholder', 'Inhalt (Markdown-Formatierung aktiviert)');
		updatePreview();
		textarea.addEventListener('input', updatePreview);
	} else {
		previewContainer.classList.add("hidden")
		textarea.removeEventListener('input', updatePreview);
		textarea.setAttribute('placeholder', 'Inhalt');
	}
}

function updatePreview() {
	var textarea = document.getElementById('content');
	var previewContent = document.getElementById('preview-content');

	if (!textarea || !previewContent) return;

	var html = textarea.value ? renderMarkdownToHTML(textarea.value) : '';
	previewContent.innerHTML = '<div class="markdown-content">' + html + '</div>';
}

function insertMarkdown(before, after) {
	var textarea = document.getElementById('content');
	if (!textarea) return;

	var start = textarea.selectionStart;
	var end = textarea.selectionEnd;
	var selectedText = textarea.value.substring(start, end);
	var newText = before + selectedText + after;

	textarea.value = textarea.value.substring(0, start) + newText + textarea.value.substring(end);

	var newPos = selectedText === '' ? start + before.length : start + before.length + selectedText.length + after.length;

	textarea.focus();
	textarea.setSelectionRange(newPos, newPos);
	updatePreview();
}

function updateExistingAttachmentsRemoval() {
	var input = document.getElementById('existing-attachments-removal');
	if (input) {
		input.value = existingAttachmentsRemoval.join(',');
	}
}

function formatFileSize(bytes) {
	if (bytes === 0) return "0 Bytes";
	var k = 1024;
	var sizes = ["Bytes", "KB", "MB", "GB"];
	var i = Math.floor(Math.log(bytes) / Math.log(k));
	return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

function onAttachments(event) {
	selectedFiles = Array.from(event.target.files);
	var previewArea = document.getElementById("file-preview");
	var container = document.getElementById("new-attachments");

	if (!previewArea || !container) return;

	container.innerHTML = "";

	if (selectedFiles.length > 0) {
		previewArea.style.display = "block";

		selectedFiles.forEach((file, index) => {
			var sizeClass = file.size > 10 * 1024 * 1024 ? "attachment-error text-red" : "muted text-sm";
			var sizeText = file.size > 10 * 1024 * 1024 ? "ZU GROSS!" : formatFileSize(file.size);

			var template = previewArea.querySelector('template[name="attachment-item"]');
			if (!template) return;

			var item = template.content.cloneNode(true);

			var nameElement = item.querySelector('.name');
			if (nameElement) nameElement.textContent = file.name;

			var sizeElement = item.querySelector('.size-text');
			if (sizeElement) {
				sizeElement.textContent = sizeText;
				sizeElement.className += ' ' + sizeClass;
			}

			var deleteBtn = item.querySelector('button.delete');
			if (deleteBtn) {
				deleteBtn.onclick = () => {
					selectedFiles.splice(index, 1);
					var fileInput = document.getElementById("attachments");
					var dt = new DataTransfer();
					selectedFiles.forEach((file) => dt.items.add(file));
					fileInput.files = dt.files;
					onAttachments(event);
				};
			}

			container.appendChild(item);
		});

		setTimeout(() => {
			previewArea.scrollIntoView({ behavior: "smooth", block: "start" });
		}, 100);
	} else {
		previewArea.style.display = "none";
	}
}

function handleDragOver(event) {
	event.preventDefault();
	event.currentTarget.classList.add('drag-over');
}

function handleDragLeave(event) {
	event.preventDefault();
	event.currentTarget.classList.remove('drag-over');
}

function handleFileDrop(event) {
	event.preventDefault();
	event.currentTarget.classList.remove('drag-over');

	var files = event.dataTransfer.files;
	if (files.length > 0) {
		var fileInput = document.getElementById('attachments');
		if (fileInput) {
			fileInput.files = files;
			onAttachments({ target: fileInput });
		}
	}
}

function deleteAttachment(attachment) {
	if (!confirm("Sind Sie sicher, dass Sie diesen Anhang löschen möchten?")) {
		return;
	}

	var attachmentItem = document.querySelector('#existing-attachments [data-attachment-path="' + attachment + '"]');

	if (attachmentItem) {
		attachmentItem.remove();
		existingAttachmentsRemoval.push(attachment);
		updateExistingAttachmentsRemoval();

		var existingAttachments = document.getElementById('existing-attachments');
		if (existingAttachments && existingAttachments.children.length === 0) {
			var detailsSection = existingAttachments.closest('details');
			if (detailsSection) {
				detailsSection.style.display = 'none';
			}
		}
	}
}
