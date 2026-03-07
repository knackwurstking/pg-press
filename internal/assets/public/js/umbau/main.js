function onToolChange(event) {
	var target = event.currentTarget;

	// Get the tool id from the target element
	//var toolID = target.value;

	// Get the format from the current selected option
	var format = "";
	var selectedOption = target.children[target.selectedIndex];
	if (selectedOption) {
		format = selectedOption.dataset.format;
	}

	// Get all the options, besides the first one for each select element (#top, #bottom)
	options = [
		...[...document.querySelectorAll("select#bottom > option")].slice(1),
		...[...document.querySelectorAll("select#top > option")].slice(1),
	]

	// Disable all options that are not compatible with the selected format
	var optionsLoopHandler = function(option) {
		option.disabled = !!format && !option.textContent.includes(format)
	};

	// Loop through all options
	options.forEach(optionsLoopHandler);
}
