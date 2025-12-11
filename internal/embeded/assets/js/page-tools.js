const idToolsFilter = "#tools-filter";
const urlSearchParamName = "tools_filter";
const storageKeyLastActiveTab = "last-active-tab";

function filterToolsList(event = null, skipHistory = false) {
	const target = event
		? event.currentTarget
		: document.querySelector(idToolsFilter);
	if (!target) return;

	const query = target.value
		.toLowerCase()
		.split(" ")
		.filter((v) => !!v);
	const targets = document.querySelectorAll(`#tools-list > .tool-item`);

	if (!skipHistory) {
		updateUrlQueryParam(query);
	}

	targets.forEach((child) => {
		child.style.display = query.every((value) =>
			child.textContent.toLowerCase().includes(value),
		)
			? "block"
			: "none";
	});
}

function initFilterInputFromQuery() {
	const urlParams = new URLSearchParams(window.location.search);
	const query = urlParams.get(urlSearchParamName);
	if (query) document.querySelector(idToolsFilter).value = query;
}

function updateUrlQueryParam(query) {
	const urlParams = new URLSearchParams(window.location.search);
	urlParams.set(urlSearchParamName, query.join(" "));
	window.history.replaceState({}, "", `?${urlParams.toString()}`);
}

function toggleTab(event) {
	document
		.querySelectorAll(".tabs > .tab")
		.forEach((tab) => tab.classList.remove("active"));

	currentTab = event.currentTarget;
	currentTab.classList.add("active");
	currentTab.dispatchEvent(new Event("loadTabContent"));

	localStorage.setItem(storageKeyLastActiveTab, currentTab.dataset.index);
}

document.addEventListener("DOMContentLoaded", () => {
	let lastActiveTab = parseInt(localStorage.getItem(storageKeyLastActiveTab));

	if (isNaN(lastActiveTab)) {
		lastActiveTab = 1;
	}

	toggleTab({
		currentTarget: document.querySelector(
			`.tabs > .tab[data-index="${lastActiveTab}"]`,
		),
	});
});
