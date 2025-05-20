// TODO: create tables
document.addEventListener("DOMContentLoaded", () => {
    const tableSearchInput =
        document.querySelector<HTMLInputElement>(`#tableSearch`)!;

    tableSearchInput.value = decodeURIComponent(queryTableSearch());

    tableSearchInput.onchange = tableSearchInputHandler as (
        ev: Event,
    ) => Promise<void>;
});

async function tableSearchInputHandler(
    ev: Event & { currentTarget: HTMLInputElement },
) {
    console.debug("search for:", ev.currentTarget.value);
    location.search = "&tableSearch=" + ev.currentTarget.value;
}

function queryTableSearch(): string {
    return (
        location.search
            .slice(1)
            .split("&")
            .find((s) => /tableSearch=/.test(s))
            ?.split("=", 2)[1] || ""
    ).replace(/\+/g, " ");
}
