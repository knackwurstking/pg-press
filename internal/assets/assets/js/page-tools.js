function filterToolsList(event = null, skipHistory = false) {
	const target = event
		? event.currentTarget
		: document.querySelector("#tools-filter");
	if (!target) return;

	const query = target.value
		.toLowerCase()
		.split(" ")
		.filter((v) => !!v);
	const targets = document.querySelectorAll(`#lists-container .tool-item`);

	if (!skipHistory) {
		updateUrlQueryParam(query);
	}

	console.debug(`Filtering tools list with query: [${query}] [skipHistory=${skipHistory}]`);

	targets.forEach((child) => {
		child.style.display = query.every((value) => {
			return child.textContent.toLowerCase().includes(value)
		})
			? "block"
			: "none";
	});
}

function initFilterInputFromQuery() {
	const urlParams = new URLSearchParams(window.location.search);
	const query = urlParams.get("tools_filter");
	if (query) document.querySelector("#tools-filter").value = query;
}

function updateUrlQueryParam(query) {
	const urlParams = new URLSearchParams(window.location.search);
	urlParams.set("tools_filter", query.join(" "));
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
