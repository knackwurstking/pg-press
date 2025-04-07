import * as ui from "ui";

import * as dialogs from "../../dialogs";
import * as globals from "../../globals";
import * as listStores from "../../list-stores/";
import * as types from "../../types";
import * as utils from "../../utils";
import * as query from "../../utils-query";
import * as create from "./create";

interface Param {
    listKey?: string;
    scrollTop?: string;
}

let cleanup: types.CleanUp[] = [];
let search: string = "";
let scrollTop: number = 0;

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const visData = globals.getVisData(param.listKey || "");
    if (!visData) {
        throw new Error(`vis data list "${param.listKey}" not found`);
    }

    // Set the app bar title
    cleanup.push(utils.setAppBarTitle(visData.title));

    if (!!param.scrollTop) {
        scrollTop = parseInt(param.scrollTop, 10);
    }

    setupAppBarEditButton(visData);
    setupAppBarAddButton(visData);
    render(visData);
}

export async function onDestroy() {
    const el = routerTargetElements();

    search = el.searchBarInput.value;
    scrollTop = el.dataList.parentElement?.scrollTop || 0;

    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function setupAppBarEditButton(visData: types.VisData) {
    const listEditButton = query.appBar_ButtonListEdit();

    listEditButton.onclick = async () => {
        const dialog = dialogs.visData(visData);

        const data = await dialog.utils!.open();
        if (!data) {
            return;
        }

        if (!dialog.utils!.validate()) {
            listEditButton.click();
            return;
        }

        try {
            const ls = listStores.get("vis-data");

            ls.replaceInStore(data, visData);
            ui.router.hash.goTo(
                {
                    listKey: ls.listKey(data),
                },
                "vis-data",
            );
        } catch (err) {
            dialog.query!.inputs.title.ariaInvalid = ""; // title input element
            alert(err);
            listEditButton.click();
            return;
        }
    };

    listEditButton.style.display = "inline-flex";

    cleanup.push(() => {
        listEditButton.style.display = "none";
        listEditButton.onclick = null;
    });
}

function setupAppBarAddButton(visData: types.VisData) {
    const addButton = query.appBar_ButtonAdd();

    addButton.onclick = async () => {
        const dialog = dialogs.visDataEntry();

        const data = await dialog.utils!.open();
        if (!data) {
            return;
        }

        // Validate data, "value" cannot be empty
        if (!dialog.utils!.validate()) {
            addButton.onclick!(new MouseEvent("click"));
            return;
        }

        visData.data.push(data);
        listStores.get("vis-data").replaceInStore(visData);
        reload();
    };

    addButton.style.display = "inline-flex";

    cleanup.push(() => {
        addButton.style.display = "none";
        addButton.onclick = null;
    });
}

function render(visData: types.VisData) {
    const el = routerTargetElements();

    el.dataList.innerHTML = "";

    // Render entries
    visData.data.forEach((entry, index) => {
        const item = create.dataItem({
            entry,
            renderTags: true,
        });
        el.dataList.appendChild(item.element);
        cleanup.push(item.destroy);

        item.element.role = "button";
        item.element.classList.add("ui-none-select");
        item.element.style.cursor = "pointer";

        item.element.oncontextmenu = (e) => {
            e.preventDefault();

            if (confirm(`You want to delete this item?\n${entry.key}\n${entry.value}`)) {
                visData.data = visData.data.filter((_e, i) => i !== index);

                const ls = listStores.get("vis-data");
                ls.replaceInStore(visData);
                reload();
            }
        };

        item.element.onclick = async () => {
            const dialog = dialogs.visDataEntry(entry);

            const newEntry = await dialog.utils!.open();

            if (!newEntry) {
                return;
            }

            if (!dialog.utils!.validate()) {
                item.element.click();
                return;
            }

            visData.data[index] = newEntry;
            listStores.get("vis-data").replaceInStore(visData);
            reload();
        };

        if (index === visData.data.length - 1 && scrollTop > 0) {
            el.dataList.parentElement!.style.scrollBehavior = "auto";
            el.dataList.parentElement!.scrollTop = scrollTop;
            el.dataList.parentElement!.style.scrollBehavior = "smooth";
        }
    });

    // Search bar
    el.searchBarInput.oninput = () => {
        const r = utils.generateRegExp(el.searchBarInput.value);
        for (const item of [...el.dataList.children]) {
            if (item.textContent === null) {
                continue;
            }

            if (!!item.textContent.replace(/(\n|\r|\s+)/g, " ").match(r)) {
                // Show
                (item as HTMLElement).style.display = "flex";
            } else {
                (item as HTMLElement).style.display = "none";
            }
        }
    };

    el.searchBarInput.value = search;
    setTimeout(() => {
        el.searchBarInput.oninput!(new Event("input"));
    });
}

async function reload() {
    await onDestroy();
    await onMount();
}

function routerTargetElements() {
    const rt = query.routerTarget();

    return {
        searchBarInput: rt.querySelector<HTMLInputElement>(`.search-bar input[type="search"]`)!,
        dataList: rt.querySelector<HTMLUListElement>(`.data-list`)!,
    };
}
