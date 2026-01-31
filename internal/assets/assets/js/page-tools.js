// ----------------------------------------------------------------------------
// This file contains the JavaScript code for the Tools page.
// It handles tab switching, filtering of tools, and preserving
// the last active tab using localStorage.
// ----------------------------------------------------------------------------

// Query parameter constants
const QUERY_TOOLS_FILTER = "tools_filter";

// ID constants
const IDS = {
	INPUT_FILTER: "tools-filter",
	LISTS_CONTAINER: "tools-container",
	//DETAILS_TOOLS: "tools-details",
	//DETAILS_CASSETTES: "cassettes-details",
};

// Class constants
const CLASSES = {
	TOOL_ITEM: "tool-item",
};

// Store details open states to preserve them during filtering
const DETAILS_OPEN_STATES = new Map();

document.addEventListener("DOMContentLoaded", () => {
	{
		// Init filter input from URL query parameter
		const urlParams = new URLSearchParams(window.location.search);
		const query = urlParams.get(QUERY_TOOLS_FILTER);
		if (query) {
			const el = document.querySelector(`#${IDS.INPUT_FILTER}`);
			if (el) {
				el.value = query;
			}
		}
	}

	{
		// Toggle last active tab
		let lastActiveTab = parseInt(localStorage.getItem("last-active-tab"));

		if (isNaN(lastActiveTab)) {
			lastActiveTab = 1;
		}

		toggleTab({
			currentTarget: document.querySelector(
				`.tabs > .tab[data-index="${lastActiveTab}"]`,
			),
		});
	}
});

// ----------------------------------------------------------------------------
// Tool Filtering
// ----------------------------------------------------------------------------

function filterToolsList(event = null, skipHistory = false) {
	const target = event
		? event.currentTarget
		: document.querySelector(`#${IDS.INPUT_FILTER}`);
	if (!target) return;

	const query = target.value
		.toLowerCase()
		.split(" ")
		.filter((v) => !!v);
	const targets = document.querySelectorAll(
		`#${IDS.LISTS_CONTAINER} .${CLASSES.TOOL_ITEM}`,
	);

	if (!skipHistory) {
		updateUrlQueryParam(query);
	}

	// Save details tag open states before filtering
	if (query.length > 0 && DETAILS_OPEN_STATES.size === 0) {
		document
			.querySelectorAll(`#${IDS.LISTS_CONTAINER} details`)
			.forEach((details) => {
				DETAILS_OPEN_STATES.set(details, details.hasAttribute("open"));
			});
	}

	if (query.length === 0) {
		targets.forEach((child) => {
			child.style.display = "block";
		});

		// Restore details tag open states
		DETAILS_OPEN_STATES.forEach((isOpen, details) => {
			if (isOpen) {
				details.setAttribute("open", "true");
			} else {
				details.removeAttribute("open");
			}
		});
		DETAILS_OPEN_STATES.clear();

		return;
	}

	targets.forEach((child) => {
		const match = query.every((value) =>
			child.textContent.toLowerCase().includes(value),
		);
		if (match) {
			child.style.display = "block";
			// If this item is inside a details tag, ensure it's open, query details tag from stack
			child.closest("details")?.setAttribute("open", "true");
			return;
		}
		child.style.display = "none";
	});
}

function updateUrlQueryParam(query) {
	const urlParams = new URLSearchParams(window.location.search);
	urlParams.set(QUERY_TOOLS_FILTER, query.join(" "));
	window.history.replaceState({}, "", `?${urlParams.toString()}`);
}

// ----------------------------------------------------------------------------
// Tab Switching
// ----------------------------------------------------------------------------

let currentTab = null;

// Tab toggle handler
function toggleTab(event) {
	document
		.querySelectorAll(".tabs .tab")
		.forEach((tab) => tab.classList.remove("active"));

	currentTab = event.currentTarget;
	currentTab.classList.add("active");
	currentTab.dispatchEvent(new Event("load-tab-content"));

	localStorage.setItem("last-active-tab", currentTab.dataset.index);
}
