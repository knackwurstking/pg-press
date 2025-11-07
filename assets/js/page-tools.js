var idToolsFilter = "tools-filter";
var idToolsList = "tools-list";
var idTabContent = "tab-content";
var urlSearchParamName = "tools_filter";
var storageKeyLastActiveTab = "last-active-tab";
var defaultTabIndex = 1;

function filterToolsList(event) {
	var target;
	if (!event) target = document.querySelector(`#${idToolsFilter}`);
	else target = event.currentTarget;

	if (!target) return;

	const urlParams = new URLSearchParams(window.location.search);
	urlParams.set(urlSearchParamName, target.value);
	window.history.replaceState({}, "", `?${urlParams.toString()}`);

	var values = target.value
		.toLowerCase()
		.split(" ")
		.filter((v) => !!v);

	var targets = document.querySelectorAll(`#${idToolsList} > .tool-item`);
	for (var child of targets) {
		var textContent = child.textContent.toLowerCase();

		var match = true;
		for (var value of values) {
			if (!textContent.includes(value)) {
				match = false;
				break;
			}
		}

		if (!match) {
			child.style.display = "none";
		} else {
			child.style.display = "block";
		}
	}
}

function initFilterInputFromQuery() {
	const urlParams = new URLSearchParams(window.location.search);
	const query = urlParams.get(urlSearchParamName);

	if (query) document.getElementById(idToolsFilter).value = query;
}

function toggleTab(event) {
	document.querySelectorAll(".tabs > .tab").forEach((tab) => {
		tab.classList.remove("active");
	});

	event.currentTarget.classList.add("active");
	event.currentTarget.dispatchEvent(new Event("loadTabContent"));

	localStorage.setItem(
		storageKeyLastActiveTab,
		event.currentTarget.dataset.index,
	);
}

// On document loaded, restore last active tab
document.addEventListener("DOMContentLoaded", () => {
	var lastActiveTab = parseInt(localStorage.getItem(storageKeyLastActiveTab));

	if (!isNaN(lastActiveTab)) {
		var tab = document.querySelector(
			`.tabs > .tab[data-index="${lastActiveTab}"]`,
		);
		if (tab) {
			toggleTab({ currentTarget: tab });
		}
	} else {
		toggleTab({
			currentTarget: document.querySelector(
				`.tabs > .tab[data-index="${defaultTabIndex}"]`,
			),
		});
	}
});
