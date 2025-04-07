import * as ui from "ui";

import * as create from "./create";
import * as globals from "../../globals";
import * as types from "../../types";
import * as query from "../../utils-query";
import * as utils from "../../utils";
import * as dialogs from "../../dialogs";
import * as listStores from "../../list-stores";

interface Param {
    listKey?: string;
    scrollTop?: string;
}

let cleanup: types.CleanUp[] = [];
let search: string = "";
let scrollTop: number = 0;

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const vis = globals.getVis(param.listKey || "");
    if (!vis) {
        throw new Error(`vis list "${param.listKey}" not found`);
    }

    // Set the app bar title
    cleanup.push(utils.setAppBarTitle(vis.title));

    if (!!param.scrollTop) {
        scrollTop = parseInt(param.scrollTop, 10);
    }

    setupAppBarEditListButton(vis);
    setupAppBarAddProductButton(vis);
    render(vis, param.listKey!);
}

export async function onDestroy() {
    const el = routerTargetElements();

    search = el.searchBarInput.value;
    scrollTop = el.products.parentElement?.scrollTop || 0;

    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function setupAppBarEditListButton(vis: types.Vis) {
    const listEditButton = query.appBar_ButtonListEdit();

    listEditButton.onclick = async () => {
        const data = await dialogs.vis(vis).utils!.open();

        if (!data) {
            return;
        }

        try {
            const ls = listStores.get("vis");
            ls.replaceInStore(data, vis);
            ui.router.hash.goTo(
                {
                    listKey: ls.listKey(data),
                },
                "vis",
            );
        } catch (err) {
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

function setupAppBarAddProductButton(vis: types.Vis) {
    const addButton = query.appBar_ButtonAdd();

    addButton.onclick = async () => {
        const data: types.Product | null = await dialogs.product().utils!.open();
        if (!data) {
            return;
        }
        vis.data.push(data);

        const ls = listStores.get("vis");
        ls.replaceInStore(vis);
        reload();
    };

    addButton.style.display = "inline-flex";

    cleanup.push(() => {
        addButton.style.display = "none";
        addButton.onclick = null;
    });
}

function render(vis: types.Vis, listKey: string) {
    const el = routerTargetElements();

    el.products.innerHTML = "";

    // Render products
    vis.data.forEach((product, index) => {
        // Seems to work just fine with more than 1000 items
        setTimeout(() => {
            const item = create.productItem({
                product,
                enableRouting: { dataIndex: index },
            });
            cleanup.push(item.destroy);
            el.products.appendChild(item.element);

            // Delete or Edit product on right click
            item.element.oncontextmenu = async (e) => {
                e.preventDefault();

                const choice = await dialogs
                    .choose(`${product.lotto}`, ["Bearbeiten", "Löschen"])
                    .utils!.open();

                switch (choice) {
                    case "Bearbeiten":
                        {
                            const data = await dialogs.product(product).utils!.open();
                            if (!data) {
                                return;
                            }
                            vis.data[index] = data;

                            const ls = listStores.get("vis");
                            ls.replaceInStore(vis);
                            reload();
                        }
                        break;

                    case "Löschen":
                        {
                            vis.data = vis.data.filter((_p, i) => i !== index);

                            const ls = listStores.get("vis");
                            ls.replaceInStore(vis);
                            reload();
                        }
                        break;
                }
            };

            if (index === vis.data.length - 1 && scrollTop > 0) {
                if (!el.products.parentElement) {
                    return;
                }

                el.products.parentElement.style.scrollBehavior = "auto";
                el.products.parentElement.scrollTop = scrollTop;
                el.products.parentElement.style.scrollBehavior = "smooth";
            }
        });
    });

    // Search bar
    el.searchBarInput.oninput = () => {
        const r = utils.generateRegExp(el.searchBarInput.value);
        for (const item of [...el.products.children]) {
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

    // Routing to product page
    el.products.onclick = (e) => {
        // Iter event path for ".product-item" element and get the "data-href" attribute
        const item = e.composedPath().find((et) => {
            return (et as HTMLElement).classList.contains("product-item");
        }) as HTMLElement | undefined;
        if (!item) {
            return;
        }

        ui.router.hash.goTo(
            {
                listKey: listKey,
                index: item.getAttribute("data-index")!,
            },
            "product",
        );
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
        products: rt.querySelector<HTMLUListElement>(`.products`)!,
    };
}
