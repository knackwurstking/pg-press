var searchTimer;

function search(event) {
	var searchValue = event.target.value.toLowerCase().trim();

	clearTimeout(searchTimer);

	searchTimer = setTimeout(function() {
		var url = new URL(window.location);
		if (searchValue) {
			url.searchParams.set("search", searchValue);
		} else {
			url.searchParams.delete("search");
		}
		history.replaceState(null, "", url);

		var searchTerms = searchValue.split(/\s+/).filter(function(term) {
			return term.length > 0;
		});

		var troubleReports = document.querySelectorAll("span.trouble-report");

		for (var i = 0; i < troubleReports.length; i++) {
			var report = troubleReports[i];
			if (searchTerms.length === 0) {
				report.style.display = "";
			} else {
				var summary = report.querySelector("summary");

				// Get content from either <pre> element or rendered markdown
				var contentText = "";
				var markdownContent = report.querySelector(".markdown-content");
				var preElement = report.querySelector("pre");

				if (markdownContent) {
					// Markdown content - get the text content of the rendered markdown
					// Use textContent to get all text, including from nested elements
					contentText = (
						markdownContent.textContent ||
						markdownContent.innerText ||
						""
					).toLowerCase();
				} else if (preElement) {
					// Non-markdown content
					contentText = preElement.textContent.toLowerCase();
				}

				var summaryText = summary
					? summary.textContent.toLowerCase()
					: "";
				var combinedText = summaryText + " " + contentText;

				var allTermsFound = true;
				for (var j = 0; j < searchTerms.length; j++) {
					if (!combinedText.includes(searchTerms[j])) {
						allTermsFound = false;
						break;
					}
				}

				report.style.display = allTermsFound ? "" : "none";
			}
		}
	}, 300);
}

function updateURLHash(event) {
	var details = event.target;
	if (details.open) {
		// Update URL hash when details is opened
		history.replaceState(null, "", "#" + details.id);
	} else {
		// Clear hash when details is closed
		history.replaceState(
			null,
			"",
			window.location.pathname + window.location.search,
		);
	}
}

window.addEventListener("beforeunload", function() {
	clearTimeout(searchTimer);
});

window.search = search;
