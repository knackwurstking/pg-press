document.addEventListener("DOMContentLoaded", () => {
    const wsStateIndicator =
        document.querySelector<HTMLElement>("#wsStateIndicator")!;

    htmx.on("htmx:wsOpen", () => {
        console.debug("ws open...");
        wsStateIndicator.innerHTML = `<span>OPEN</span>`;
    });

    htmx.on("htmx:wsClose", () => {
        console.debug("ws close...");
        wsStateIndicator.innerHTML = `<span>CLOSE</span>`;
    });

    htmx.on("htmx:wsError", () => {
        console.debug("ws error...");
        wsStateIndicator.innerHTML = `<span>ERROR</span>`;
    });
});
