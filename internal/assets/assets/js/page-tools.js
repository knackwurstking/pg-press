function filterToolsList(event = null, skipHistory = false) {
	const target = event
		? event.currentTarget
		: document.querySelector("#tools-filter");
	if (!target) return;

	const query = target.value
		.toLowerCase()
		.trim();
	const targets = document.querySelectorAll(`#lists-container .tool-item`);

	if (!skipHistory) {
		updateUrlQueryParam(query);
	}

	console.debug(`Filtering tools list with query: [${query}] [skipHistory=${skipHistory}]`);

	// If query is empty, show all items and close details
	if (!query) {
		targets.forEach((child) => {
			child.style.display = "block";
		});

		// Close all details when no search term
		document.querySelectorAll("#lists-container details").forEach(detail => {
			detail.open = false;
		});

		// Update counters when clearing filter
		updateCounters();
		return;
	}

	// Parse query to support field-specific filtering (like "type:FC" or "code:G01")
	const parsedQuery = parseFilterQuery(query);

	targets.forEach((child) => {
		const isMatch = matchesFilter(child, parsedQuery);
		child.style.display = isMatch ? "block" : "none";
	});

	// Auto-open details that contain matching items
	autoOpenDetails();

	// Update counters after filtering
	updateCounters();
}

function autoOpenDetails() {
	const details = document.querySelectorAll("#lists-container details");

	details.forEach(detail => {
		// Check if any child tool items are visible
		const toolItems = detail.querySelectorAll(".tool-item");
		let hasVisibleItems = false;

		toolItems.forEach(item => {
			if (item.style.display !== "none") {
				hasVisibleItems = true;
			}
		});

		// Open the details if it has visible items
		detail.open = hasVisibleItems;
	});
}

function parseFilterQuery(query) {
	// Split by spaces but keep quoted strings together
	const tokens = [];
	let currentToken = '';
	let inQuotes = false;
	let quoteChar = '';

	for (let i = 0; i < query.length; i++) {
		const char = query[i];

		if (char === '"' || char === "'") {
			if (!inQuotes) {
				inQuotes = true;
				quoteChar = char;
			} else if (quoteChar === char) {
				inQuotes = false;
				quoteChar = '';
				tokens.push(currentToken);
				currentToken = '';
			} else {
				currentToken += char;
			}
		} else if (char === ' ' && !inQuotes) {
			if (currentToken) {
				tokens.push(currentToken);
				currentToken = '';
			}
		} else {
			currentToken += char;
		}
	}

	if (currentToken) {
		tokens.push(currentToken);
	}

	// Parse tokens into field:value pairs
	const fields = {};
	const generalTerms = [];

	for (const token of tokens) {
		const colonIndex = token.indexOf(':');
		if (colonIndex !== -1) {
			const field = token.substring(0, colonIndex).toLowerCase();
			const value = token.substring(colonIndex + 1);
			fields[field] = value;
		} else {
			generalTerms.push(token);
		}
	}

	return {
		fields: fields,
		general: generalTerms
	};
}

function matchesFilter(element, parsedQuery) {
	// Get the text content of the element for general term matching
	const elementText = element.textContent.toLowerCase();

	// Check field-specific filters first
	for (const [field, value] of Object.entries(parsedQuery.fields)) {
		if (!matchesField(element, field, value)) {
			return false;
		}
	}

	// Check general terms (all must match)
	if (parsedQuery.general.length > 0) {
		return parsedQuery.general.every(term => {
			return elementText.includes(term);
		});
	}

	return true;
}

function matchesField(element, field, value) {
	// Extract tool/cassette data from the element's HTML structure
	const toolData = extractToolData(element);

	// Check if the element has the required field data
	if (toolData && toolData[field]) {
		const fieldData = toolData[field].toLowerCase();
		return fieldData.includes(value.toLowerCase());
	}

	// If the field is not found in our data structure, fall back to text matching
	const elementText = element.textContent.toLowerCase();
	return elementText.includes(value.toLowerCase());
}

function extractToolData(element) {
	// This function attempts to extract tool/cassette information from the rendered element
	// The actual implementation may need to be enhanced based on the specific HTML structure
	try {
		// Get the tool ID from the element ID (e.g., tool-123 or cassette-456)
		const id = element.id;
		if (!id) return null;

		const toolMatch = id.match(/tool-(\d+)/);
		const cassetteMatch = id.match(/cassette-(\d+)/);

		// Return structure with basic information
		return {
			id: id,
			isTool: !!toolMatch,
			isCassette: !!cassetteMatch,
			// Add more fields that can be extracted from the element
			type: null,  // Will be filled by other methods
			code: null,  // Will be filled by other methods
			width: null, // Will be filled by other methods
			height: null // Will be filled by other methods
		};
	} catch (e) {
		console.warn("Error extracting tool data from element:", e);
		return null;
	}
}

// Optimized version for better performance
function smartFilterToolsList(event = null, skipHistory = false) {
	const target = event
		? event.currentTarget
		: document.querySelector("#tools-filter");
	if (!target) return;

	const query = target.value
		.toLowerCase()
		.trim();
	const targets = document.querySelectorAll(`#lists-container .tool-item`);

	if (!skipHistory) {
		updateUrlQueryParam(query);
	}

	console.debug(`Filtering tools list with smart query: [${query}] [skipHistory=${skipHistory}]`);

	// If query is empty, show all items
	if (!query) {
		targets.forEach((child) => {
			child.style.display = "block";
		});
		return;
	}

	// Simple and smarter filter logic - match ANY of the terms instead of ALL
	const terms = query.split(/\s+/).filter(term => term.length > 0);

	targets.forEach((child) => {
		// Get text content of the element
		const elementText = child.textContent.toLowerCase();

		// Match ANY of the terms (smarter than requiring ALL terms)
		const isMatch = terms.some(term => elementText.includes(term));

		child.style.display = isMatch ? "block" : "none";
	});
}

function initFilterInputFromQuery() {
	const urlParams = new URLSearchParams(window.location.search);
	const query = urlParams.get("tools_filter");
	if (query) document.querySelector("#tools-filter").value = query;
}

function updateUrlQueryParam(query) {
	const urlParams = new URLSearchParams(window.location.search);
	urlParams.set("tools_filter", query);
	window.history.replaceState({}, "", `?${urlParams.toString()}`);
}

function toggleTab(event) {
	document
		.querySelectorAll(".tabs .tab")
		.forEach((tab) => tab.classList.remove("active"));

	currentTab = event.currentTarget;
	currentTab.classList.add("active");
	currentTab.dispatchEvent(new Event("loadTabContent"));

	localStorage.setItem("last-active-tab", currentTab.dataset.index);
}

function updateCounters() {
	// Update tools counter
	const toolsDetails = document.querySelector('#tools-details');
	const toolsItems = document.querySelectorAll('#tools-details .tool-item');
	const visibleToolsCount = Array.from(toolsItems).filter(item => item.style.display !== 'none').length;

	if (toolsDetails) {
		const toolsSummary = toolsDetails.querySelector('summary');
		if (toolsSummary) {
			toolsSummary.textContent = visibleToolsCount === 1
				? '1 Werkzeug'
				: `${visibleToolsCount} Werkzeuge`;
		}
	} else {
		console.warn('Tools details element not found for counter update.');
	}

	// Update cassettes counter
	const cassettesDetails = document.querySelector('#cassettes-details');
	const cassettesItems = document.querySelectorAll('#cassettes-details .tool-item');
	const visibleCassettesCount = Array.from(cassettesItems).filter(item => item.style.display !== 'none').length;

	if (cassettesDetails) {
		const cassettesSummary = cassettesDetails.querySelector('summary');
		if (cassettesSummary) {
			cassettesSummary.textContent = visibleCassettesCount === 1
				? '1 Kassette'
				: `${visibleCassettesCount} Kassetten`;
		}
	} else {
		console.warn('Cassettes details element not found for counter update.');
	}
}

document.addEventListener("DOMContentLoaded", () => {
	let lastActiveTab = parseInt(localStorage.getItem("last-active-tab"));

	if (isNaN(lastActiveTab)) {
		lastActiveTab = 1;
	}

	toggleTab({
		currentTarget: document.querySelector(
			`.tabs > .tab[data-index="${lastActiveTab}"]`,
		),
	});
});
