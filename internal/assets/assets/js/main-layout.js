var globalHxTrigger = "pageLoaded";

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
	function() {
		updateDataTheme();
	},
);

// TODO: Can be removed? Not sure if needed anymore, but we will see.
//document.addEventListener("DOMContentLoaded", function() {
//	//// Debounce function to prevent rapid reloads
//	function debounce(func, wait) {
//		let timeout;
//		return function executedFunction(...args) {
//			const later = function() {
//				clearTimeout(timeout);
//				func(...args);
//			};
//			clearTimeout(timeout);
//			timeout = setTimeout(later, wait);
//		};
//	}
//
//	// Debounced reload function
//	const debouncedReload = debounce(function() {
//		if (document.visibilityState === "visible") {
//			console.log("Page became visible - reloading HTMX sections");
//			document.body.dispatchEvent(new CustomEvent(globalHxTrigger));
//		}
//	}, 500);
//
//	// Listen for visibility changes
//	window.addEventListener("visibilitychange", debouncedReload);
//
//	// Add manual refresh functionality
//	window.refreshPressSections = function() {
//		console.log("Manual refresh triggered");
//		document.body.dispatchEvent(new CustomEvent(globalHxTrigger));
//	};
//});
