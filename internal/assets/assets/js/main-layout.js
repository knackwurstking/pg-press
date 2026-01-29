// For the ui.min.css i need to set the data-theme to light/dark
function updateDataTheme() {
	const themeColorMeta = document.getElementById("theme-color-meta");
	if (matchMedia("(prefers-color-scheme: dark)").matches) {
		document.querySelector("html").setAttribute("data-theme", "dark");
		// Dark theme background color
		if (themeColorMeta) themeColorMeta.setAttribute("content", "#1d2021");
	} else {
		document.querySelector("html").setAttribute("data-theme", "light");
		// Light theme background color
		if (themeColorMeta) themeColorMeta.setAttribute("content", "#f9f5d7");
	}
}

updateDataTheme();

matchMedia("(prefers-color-scheme: dark)").addEventListener(
	"change",
	updateDataTheme,
);

document.addEventListener("DOMContentLoaded", function () {
	window.hxTriggers = window.hxTriggers || [];

	// Debounce function to prevent rapid reloads
	function debounce(func, wait) {
		let timeout;
		return function executedFunction(...args) {
			const later = function () {
				clearTimeout(timeout);
				func(...args);
			};
			clearTimeout(timeout);
			timeout = setTimeout(later, wait);
		};
	}

	// Listen for visibility changes
	window.addEventListener(
		"visibilitychange",
		debounce(function () {
			if (document.visibilityState === "visible") {
				console.log("Page became visible - reloading HTMX sections");
				if (window.hxTriggers.length > 0) {
					console.debug("Triggers: ", window.hxTriggers);
					window.hxTriggers.forEach(function (trigger) {
						// Add error handling for custom event dispatch
						try {
							document.body.dispatchEvent(
								new CustomEvent(trigger),
							);
						} catch (error) {
							console.warn(
								"Failed to dispatch HTMX trigger:",
								trigger,
								error,
							);
						}
					});
				}
			}
		}, 500),
	);
});

window.setHxTrigger = function (...triggers) {
	triggers.forEach(function (trigger) {
		if (!window.hxTriggers.includes(trigger)) {
			window.hxTriggers.push(trigger);
		}
	});
};
