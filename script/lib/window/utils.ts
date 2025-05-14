export function create(): Utils {
    return {
        setupAppBarItems,
        setOnlineIndicatorState,
        registerServiceWorker,
    };
}

function setupAppBarItems(...itemNames: AppBarItemName[]): AppBarItems {
    const enabledItems: AppBarItems = {};

    const items = document.querySelectorAll<HTMLElement>(
        `.ui-app-bar [data-name]`,
    );
    let match = false;
    for (const item of items) {
        const dataName = (item.getAttribute("data-name") ||
            "") as AppBarItemName;

        match = false;
        for (const name of itemNames) {
            if (name === dataName) {
                match = true;
            }
        }

        if (match) {
            // Enable
            item.style.display = "inline-flex";
            enabledItems[dataName] = item;
        } else {
            // Disable
            item.style.display = "none";
        }
    }

    return enabledItems;
}

function setOnlineIndicatorState(state: boolean): void {
    const el = document.querySelector(`.online-indicator`);
    if (!el) return;

    if (state) {
        el.setAttribute(`data-state`, "online");
    } else {
        el.setAttribute(`data-state`, "offline");
    }
}

function registerServiceWorker(): void {
    // Check if the browser supports service workers, otherwise abort.
    if (!("serviceWorker" in navigator)) {
        console.warn("Browser doesn't support service workers");
        return;
    }

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
}
