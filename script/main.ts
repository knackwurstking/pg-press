/*
 * Register service worker
 */

// Check if the browser supports service workers, otherwise abort.
if ("serviceWorker" in navigator) {
    window.addEventListener("pageshow", function () {
        navigator.serviceWorker
            .register(process.env.SERVER_PATH_PREFIX + "/service-worker.js")
            .then(function (reg) {
                console.info("Service worker registered", reg);
            })
            .catch(function (err) {
                console.error("Service worker registration failed:", err);
            });
    });
} else {
    console.warn("Browser doesn't support service workers");
}

// TODO: global WebSocket (window.ws)
