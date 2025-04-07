import * as ui from "ui";

import * as globals from "../../globals";
import * as types from "../../types";
import * as utils from "../../utils";
import * as query from "../../utils-query";
import * as createAlertLists from "../alert-lists/create";
import * as create from "./create";

interface Param {
    listKey?: string;
    index?: string;
}

let cleanup: types.CleanUp[] = [];

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const alert = globals.getAlert(param.listKey || "", parseInt(param.index || "-1", 10));
    if (!alert) {
        throw `alert not found: listKey=${param.listKey}, index=${param.index}`;
    }

    cleanup.push(
        utils.setAppBarTitle(
            alert.from === alert.to ? `${alert.from}` : `${alert.from}..${alert.to}`,
        ),
    );

    // Enable back button
    setupAppBarBackButton();
    render(alert);
}

export async function onDestroy() {
    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function setupAppBarBackButton() {
    const backButton = query.appBar_ButtonBack();
    backButton.style.display = "inline-flex";
    cleanup.push(() => {
        backButton.style.display = "none";
    });
}

function render(alert: types.Alert) {
    const el = routerTargetElements();

    // Reset, just in case
    el.alert.innerHTML = "";
    el.desc.innerHTML = "";

    const alertItem = createAlertLists.alertItem({ alert: alert });

    el.alert?.appendChild(alertItem.element);
    cleanup.push(alertItem.destroy);

    const descItem = create.alertDescription({ alert });
    el.desc.appendChild(descItem.element);
    cleanup.push(descItem.destroy);
}

function routerTargetElements() {
    const routerTarget = query.routerTarget();

    return {
        alert: routerTarget.querySelector(`.alert`)!,
        desc: routerTarget.querySelector(`.desc`)!,
    };
}
