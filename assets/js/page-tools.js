var IDToolsFilter = "tools-filter";
var IDSectionToolsList = "section-tools-list";

function scrollFilterInputIntoView(event) {
	console.debug("Scrolling filter input");

	var section = event.currentTarget.closest(`#${IDSectionToolsList}`);
	setTimeout(function () {
		section.scrollIntoView({ behavior: "smooth", block: "start" });
	}, 100);
}

function unscrollFilterInputIntoView(event) {
	console.debug("Unscrolling filter input");

	var section = event.currentTarget.closest(`#${IDSectionToolsList}`);
	section.style.height = "fit-content";
}

function filter(value) {
	for (var el of document.querySelectorAll(`.all-tools ul > *`)) {
		// Get the text content of the element
		var textContent = el.textContent.toLowerCase();

		// Generate special regexp from search
		var valueSplit = value.toLowerCase().split(" ");

		// Check and Set "block" or "none"
		var shouldNotHide = true;
		for (var char of valueSplit) {
			if (char === "") continue;

			if (!textContent.includes(char)) {
				shouldNotHide = false;
				break;
			}
		}

		el.style.display = shouldNotHide ? "block" : "none";
	}
}

function filterToolsList(event) {
	var input = document.querySelector(`input#${IDToolsFilter}`);

	var params = new URLSearchParams(location.search);
	if (input.value !== "") {
		params.set("tools_filter", input.value);
	} else {
		params.delete("tools_filter");
	}

	filter(input.value);

	// Update browser history to include filter parameters
	var newUrl =
		window.location.pathname +
		(params.toString() ? "?" + params.toString() : "");
	window.history.replaceState({}, "", newUrl);
}

document.addEventListener("DOMContentLoaded", function () {
	var params = new URLSearchParams(window.location.search);
	if (params.toString()) {
		var toolsSection = document.getElementById(`${IDSectionToolsList}`);
		if (toolsSection) {
			toolsSection.open = true;
			htmx.trigger(toolsSection, "toggle");
		}
	}
});
