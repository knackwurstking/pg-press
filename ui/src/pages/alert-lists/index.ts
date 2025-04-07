import * as ui from "ui";

import * as globals from "../../globals";
import * as types from "../../types";
import * as utils from "../../utils";
import * as query from "../../utils-query";
import * as create from "./create";

interface Param {
    listKey?: string;
}

let cleanup: types.CleanUp[] = [];

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const list = globals.getAlertList(param.listKey || "");
    if (!list) {
        throw new Error(`alert list "${param.listKey}" not found`);
    }

    cleanup.push(utils.setAppBarTitle(list.title));

    render(list.data, param.listKey || "");
}

export async function onDestroy() {
    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function render(alerts: types.Alert[], listKey: string) {
    const el = routerTargetElements();

    alerts.forEach((alert, i) => {
        setTimeout(() => {
            const item = create.alertItem({
                alert: alert,
                enableRouting: {
                    alertIndex: i,
                    listKey,
                },
            });
            cleanup.push(item.destroy);
            el.alerts.appendChild(item.element);
        });
    });

    el.searchBarInput.oninput = () => {
        const r = utils.generateRegExp(el.searchBarInput.value);
        for (const item of [...el.alerts.children]) {
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
}

function routerTargetElements() {
    const routerTarget = query.routerTarget();

    return {
        searchBarInput: routerTarget.querySelector<HTMLInputElement>(
            `.search-bar input[type="search"]`,
        )!,

        alerts: routerTarget.querySelector<HTMLUListElement>(`.alerts`)!,
    };
}
