import * as ui from "ui";

import * as globals from "../../globals";
import * as query from "../../utils-query";
import * as types from "../../types";
import * as utils from "../../utils";
import * as listStores from "../../list-stores";
import * as visCreate from "../vis/create";

interface Param {
    listKey?: string;
}

let cleanup: types.CleanUp[] = [];

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const list = globals.getVisBookmarks(param.listKey || "");
    if (!list) {
        throw new Error(`vis bookmarks "${param.listKey}" not found`);
    }

    // Set the app bar title
    cleanup.push(utils.setAppBarTitle(list.title));

    render(list);
}

export async function onDestroy() {
    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function render(list: types.Bookmarks) {
    const el = routerTargetElements();

    el.products.innerHTML = "";

    // Render product items
    list.data.forEach((product) => {
        // Get the product from the vis lists if possible (keys: "lotto" & "name")
        const { visProduct, visListKey, visDataIndex } = searchVisForProduct(product);

        const item = visCreate.productItem({
            product: visProduct,
            enableRouting: { dataIndex: visDataIndex },
        });

        item.element.setAttribute("data-json", JSON.stringify(product));
        item.element.onclick = () => {
            ui.router.hash.goTo(
                {
                    listKey: visListKey,
                    index: visDataIndex.toString(),
                },
                "product",
            );
        };

        el.products.appendChild(item.element);
        cleanup.push(item.destroy);
    });

    // Make it draggable
    ui.draggable.createMobile(el.products, {
        onDragEnd: async () => {
            const newData: types.Product[] = [];

            // Update bookmarks data and reload
            Array.from(el.products.children).forEach((child) => {
                const product: types.Product = JSON.parse(child.getAttribute("data-json")!);
                newData.push(product);
            });

            const ls = listStores.get("vis-bookmarks");
            list.data = newData;
            ls.replaceInStore(list);

            await reload();
        },
    });
}

function routerTargetElements() {
    return {
        products: query.routerTarget().querySelector<HTMLUListElement>(`.products`)!,
    };
}

async function reload() {
    await onDestroy();
    await onMount();
}

function searchVisForProduct(product: types.Product) {
    const ls = listStores.get("vis");

    for (const list of globals.store.get("vis")!.lists) {
        let visDataIndex = 0;
        for (const visProduct of list.data) {
            if (visProduct.lotto === product.lotto && visProduct.name === product.name) {
                return { visProduct, visListKey: ls.listKey(list), visDataIndex };
            }

            visDataIndex++;
        }
    }

    return { visProduct: product, visListKey: "", visDataIndex: -1 };
}
