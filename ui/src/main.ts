import "./bootstrap-icons.css";

import { App } from "@capacitor/app";
import * as ui from "ui";
import { registerSW } from "virtual:pwa-register";

import * as dialogs from "./dialogs";
import * as drawer from "./drawer";
import * as globals from "./globals";
import * as listStores from "./list-stores";
import * as pages from "./pages";
import * as types from "./types";
import * as utils from "./utils";
import * as query from "./utils-query";

if (process.env.CAPACITOR) {
    App.addListener("backButton", ({ canGoBack }) => {
        if (!!document.querySelector(`dialog[open]`)) {
            return;
        }

        if (!canGoBack) {
            App.exitApp();
        } else {
            window.history.back();
        }
    });
}

if (process.env.PWA) {
    const updateSW = registerSW({
        async onNeedRefresh() {
            if (confirm(`Update verfügbar. Zum Aktualisieren bestätigen.`)) {
                await updateSW();
            }
        },
    });
}

// Initialize AppBar

query.appBar_ButtonOpenDrawer().onclick = () => drawer.utils.open();
query.drawerBackdrop().onclick = () => drawer.utils.close();

// Initialize Drawer

const drawerGistIDsButton = query.drawerGistIDsButton();
drawerGistIDsButton.onclick = () => {
    ui.router.hash.goTo(null, "");
    drawer.utils.close();
};

// Initialize Main Event Handlers (store)

// Render drawer alert-lists items
{
    let cleanup: (() => void)[] = [];
    globals.store.listen(
        "alert-lists",
        async (data) => {
            cleanup.forEach((fn) => fn());
            cleanup = [];

            const group = query.drawerGroup("alert-lists");

            // Render lists
            {
                group.items.innerHTML = "";
                for (const list of data.lists) {
                    const item = drawer.create.alertListItem({ data: list });
                    cleanup.push(item.destroy);
                    group.items.appendChild(item.element);
                }
            }

            // Initialize action button "download" - Download data packed in a handy zip archive
            group.actions.download!.onclick = async () => {
                try {
                    if (confirm(`Download ZIP?`)) {
                        await utils.downloadZIP("alert-lists");
                    }
                } catch (err) {
                    alert(err);
                }
            };

            // Initialize action "import-from-file"
            group.actions.importFromFile!.onclick = () => {
                utils.importFromFile(".json", "alert-lists");
            };
        },
        true,
    );
}

// Render drawer metal-sheets items
{
    let cleanup: (() => void)[] = [];
    globals.store.listen(
        "metal-sheets",
        async (data) => {
            cleanup.forEach((fn) => fn());
            cleanup = [];

            // Render
            const group = query.drawerGroup("metal-sheets");

            // Render active lists first, then the rest
            {
                // Sort active lists by press
                let activeLists: types.MetalSheet[] = [];
                const rest: types.MetalSheet[] = [];
                for (const list of data.lists) {
                    if (list.data.press > -1) {
                        activeLists.push(list);
                    } else {
                        rest.push(list);
                    }
                }

                // Sort active lists by press from low to high
                activeLists = activeLists.sort((a, b) =>
                    a.data.press > b.data.press ? 1 : -1,
                );

                group.items.innerHTML = "";
                for (const list of [...activeLists, ...rest]) {
                    const item = drawer.create.metalSheetItem({ data: list });
                    cleanup.push(item.destroy);
                    group.items.appendChild(item.element);
                }
            }

            // Initialize action button "download" - Download data packed in a handy zip archive
            group.actions.download!.onclick = async () => {
                try {
                    if (confirm(`Download ZIP?`)) {
                        await utils.downloadZIP("metal-sheets");
                    }
                } catch (err) {
                    alert(err);
                }
            };

            // Initialize action "import-from-file"
            group.actions.importFromFile!.onclick = () => {
                utils.importFromFile(".json", "metal-sheets");
            };

            // This will open the metal-sheet dialog
            group.actions.add!.onclick = async () => {
                const dialog = dialogs.metalSheet();

                const data = await dialog.utils!.open();
                if (!data) return;

                // Format should not be empty
                if (!dialog.utils!.validate()) {
                    group.actions.add!.click();
                    return;
                }

                try {
                    listStores.get("metal-sheets").addToStore([data]);
                } catch (err) {
                    alert(err);
                    group.actions.add!.click();
                    return;
                }
            };
        },
        true,
    );
}

// Render drawer vis items
{
    let cleanup: (() => void)[] = [];
    globals.store.listen(
        "vis",
        async (data) => {
            cleanup.forEach((fn) => fn());
            cleanup = [];

            const group = query.drawerGroup("vis");

            // Render lists
            {
                group.items.innerHTML = "";
                for (const list of data.lists) {
                    const item = drawer.create.visItem({ data: list });
                    cleanup.push(item.destroy);
                    group.items.appendChild(item.element);
                }
            }

            // Initialize action button "download" - Download data packed in a handy zip archive
            group.actions.download!.onclick = async () => {
                try {
                    if (confirm(`Download ZIP?`)) {
                        await utils.downloadZIP("vis");
                    }
                } catch (err) {
                    alert(err);
                }
            };

            // Initialize action button "import-from-file" - Open file picker
            group.actions.importFromFile!.onclick = () => {
                utils.importFromFile(".txt,.json", "vis");
            };

            // Initialize action button "add" - Create a new vis list
            group.actions.add!.onclick = async () => {
                // Open dialog choosing load from file or create new vis list
                const data = await dialogs.vis().utils!.open();

                if (!data) {
                    return;
                }

                const ls = listStores.get("vis");
                try {
                    ls.addToStore([data]);
                } catch (err) {
                    alert(err);
                    group.actions.add!.click();
                    return;
                }
            };
        },
        true,
    );
}

// Render drawer vis bookmarks items
{
    let cleanup: (() => void)[] = [];
    globals.store.listen(
        "vis-bookmarks",
        async (data) => {
            cleanup.forEach((fn) => fn());
            cleanup = [];

            const group = query.drawerGroup("vis-bookmarks");

            // Render lists
            {
                group.items.innerHTML = "";
                for (const list of data.lists) {
                    const item = drawer.create.visBookmarkItem({ data: list });

                    if (!list.allowDeletion) {
                        item.query!.deleteButton().setAttribute("disabled", "");
                    }

                    group.items.appendChild(item.element);
                    cleanup.push(item.destroy);
                }
            }
        },
        true,
    );
}

// Render drawer vis data items
{
    let cleanup: (() => void)[] = [];
    globals.store.listen(
        "vis-data",
        async (data) => {
            cleanup.forEach((fn) => fn());
            cleanup = [];

            const group = query.drawerGroup("vis-data");

            // Render lists
            {
                group.items.innerHTML = "";
                for (const list of data.lists) {
                    const item = drawer.create.visDataItem({ data: list });
                    group.items.appendChild(item.element);
                    cleanup.push(item.destroy);
                }
            }

            // Initialize action button "download" - Download data packed in a handy zip archive
            group.actions.download!.onclick = async () => {
                try {
                    if (confirm(`Download ZIP?`)) {
                        await utils.downloadZIP("vis-data");
                    }
                } catch (err) {
                    alert(err);
                }
            };

            // Initialize action button "import-from-file" - Open file picker
            group.actions.importFromFile!.onclick = () => {
                utils.importFromFile(".json", "vis-data");
            };

            // Initialize action button "add" - Create a new vis list
            group.actions.add!.onclick = async () => {
                const dialog = dialogs.visData();

                const data = await dialog.utils!.open();
                if (!data) {
                    return;
                }

                if (!dialog.utils!.validate()) {
                    group.actions.add!.click();
                    return;
                }

                try {
                    listStores.get("vis-data").addToStore([data]);
                } catch (err) {
                    dialog.query!.inputs.title.ariaInvalid = ""; // title input element
                    alert(err);
                    group.actions.add!.click();
                    return;
                }
            };
        },
        true,
    );
}

// Render drawer special items
{
    let cleanup: (() => void)[] = [];
    globals.store.listen(
        "special",
        async (data) => {
            cleanup.forEach((fn) => fn());
            cleanup = [];

            const group = query.drawerGroup("special");

            // Render lists
            {
                group.items.innerHTML = "";
                for (const list of data.lists) {
                    // Type check of special flakes
                    const item = drawer.create.specialItem({ data: list });
                    group.items.appendChild(item.element);
                    cleanup.push(item.destroy);
                }
            }

            // Initialize action button "download" - Download data packed in a handy zip archive
            group.actions.download!.onclick = async () => {
                try {
                    if (confirm(`Download ZIP?`)) {
                        await utils.downloadZIP("special");
                    }
                } catch (err) {
                    alert(err);
                }
            };

            // Initialize action button "import-from-file" - Open file picker
            group.actions.importFromFile!.onclick = () => {
                utils.importFromFile(".json", "special");
            };
        },
        true,
    );
}

// Initialize Router

ui.router.hash.init(query.routerTarget(), {
    "/": {
        title: "VIS | Gist IDs",
        template: {
            selector: `template[name="gist-ids"]`,
            onMount() {
                pages.gistIDs.onMount();
            },
            onDestroy() {
                pages.gistIDs.onDestroy();
            },
        },
    },

    "alert-lists": {
        title: "VIS | Alarm Listen",
        template: {
            selector: `template[name="alert-lists"]`,
            onMount() {
                pages.alertLists.onMount();
            },
            onDestroy() {
                pages.alertLists.onDestroy();
            },
        },
    },

    alert: {
        title: "VIS | Alarm",
        template: {
            selector: `template[name="alert"]`,
            onMount() {
                pages.alert.onMount();
            },
            onDestroy() {
                pages.alert.onDestroy();
            },
        },
    },

    "metal-sheets": {
        title: "VIS | Blech Listen",
        template: {
            selector: `template[name="metal-sheets"]`,
            onMount() {
                pages.metalSheets.onMount();
            },
            onDestroy() {
                pages.metalSheets.onDestroy();
            },
        },
    },

    vis: {
        title: "VIS",
        template: {
            selector: `template[name="vis"]`,
            onMount() {
                pages.vis.onMount();
            },
            onDestroy() {
                pages.vis.onDestroy();
            },
        },
    },

    product: {
        title: "VIS | Produkt",
        template: {
            selector: `template[name="product"]`,
            onMount() {
                pages.product.onMount();
            },
            onDestroy() {
                pages.product.onDestroy();
            },
        },
    },

    "vis-bookmarks": {
        title: "VIS | Bookmarks",
        template: {
            selector: `template[name="vis-bookmarks"]`,
            onMount() {
                pages.visBookmarks.onMount();
            },
            onDestroy() {
                pages.visBookmarks.onDestroy();
            },
        },
    },

    "vis-data": {
        title: "VIS | Data",
        template: {
            selector: `template[name="vis-data"]`,
            onMount() {
                pages.visData.onMount();
            },
            onDestroy() {
                pages.visData.onDestroy();
            },
        },
    },

    "special-flakes": {
        title: "VIS | Flakes",
        template: {
            selector: `template[name="special-flakes"]`,
            onMount() {
                pages.special.flakes.onMount();
            },
            onDestroy() {
                pages.special.flakes.onDestroy();
            },
        },
    },
});

// Event Handlers

window.onhashchange = () => {
    drawer.utils.close();
};
