function toggleTab(event) {
	document.querySelectorAll(".tabs > .tab").forEach((tab) => {
		tab.classList.remove("active");
	});

	event.currentTarget.classList.add("active");

	var spinnerContainer = document.querySelector(
		".tab-content > .spinner-container",
	);
	if (spinnerContainer) {
		spinnerContainer.style.display = "block";
	}

	localStorage.setItem("last-active-tab", event.currentTarget.dataset.index);
}

// On document loaded, restore last active tab
document.addEventListener("DOMContentLoaded", () => {
	var lastActiveTab = parseInt(localStorage.getItem("last-active-tab"));

	if (!isNaN(lastActiveTab)) {
		var tab = document.querySelector(
			`.tabs > .tab[data-index="${lastActiveTab}"]`,
		);
		if (tab) {
			toggleTab({ currentTarget: tab });
		}
	}
});
