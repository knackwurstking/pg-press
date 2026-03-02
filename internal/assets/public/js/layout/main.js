function updateDataTheme() {
	const themeColorMeta = document.getElementById("theme-color-meta");
	if (matchMedia("(prefers-color-scheme: dark)").matches) {
		document.querySelector("html").classList.add("dark");
		// Dark theme background color
		if (themeColorMeta) themeColorMeta.setAttribute("content", "#1d2021");
	} else {
		document.querySelector("html").classList.add("dark");
		// Light theme background color
		if (themeColorMeta) themeColorMeta.setAttribute("content", "#f9f5d7");
	}
}

updateDataTheme();

matchMedia("(prefers-color-scheme: dark)").addEventListener(
	"change",
	updateDataTheme,
);

document.addEventListener("DOMContentLoaded", function() {
	window.triggers = window.triggers || [];

	// Debounce function to prevent rapid reloads
	function debounce(func, wait) {
		let timeout;
		return function executedFunction(...args) {
			const later = function() {
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
		debounce(function() {
			if (document.visibilityState === "visible") {
				// Do nothing if any dialog is open
				var dialogs = document.querySelectorAll("dialog[open]")
				if (dialogs.length > 0) {
					console.debug("Dialogs are open, skipping HTMX reload");
					return;
				}
				dialogs = document.querySelectorAll(`[data-tui-dialog-open="true"]`)
				if (dialogs.length > 0) {
					console.debug("TUI Dialogs are open, skipping HTMX reload");
					return;
				}

				// Reload HTMX sections by dispatching custom events
				console.debug("Page became visible - reloading HTMX sections");
				if (window.triggers.length > 0) {
					console.debug("Triggers: ", window.triggers);
					window.triggers.forEach(function(trigger) {
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

window.setTriggers = function(...triggers) {
	triggers.forEach(function(trigger) {
		if (!window.triggers.includes(trigger)) {
			window.triggers.push(trigger);
		}
	});
};
