function onToolChange() {
	// Get all the options, besides the first one for each select element (#top, #bottom)
	var options = [...document.querySelectorAll(".tool-select-item")];

	// Find the selected item
	var selectedItem = null;
	for (var option of options) {
		if (option.dataset.tuiSelectboxSelected === "true") {
			selectedItem = option;
			break;
		}
	}

	// Get the format from the selected item
	var format;
	if (selectedItem) {
		format = selectedItem.dataset.format;
	}

	// Disable all options that are not compatible with the selected format
	var disabledClassNames = ["pointer-events-none", "opacity-50"]
	var optionsLoopHandler = function(option) {
		if (!!format && !option.textContent.includes(format)) {
			option.classList.add(...disabledClassNames);
		} else {
			option.classList.remove(...disabledClassNames);
		}
	}

	// Loop through all options
	options.forEach(optionsLoopHandler);
}
