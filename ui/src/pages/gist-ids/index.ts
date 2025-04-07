import * as query from "../../utils-query";
import * as create from "./create";
import * as utils from "../../utils";
import * as types from "../../types";

let cleanup: types.CleanUp[] = [];

export async function onMount() {
    cleanup.push(utils.setAppBarTitle("Gist IDs"));

    setupAppBarDevModeButton();
    render();
}

export async function onDestroy() {
    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function setupAppBarDevModeButton() {
    const dataBaseButton = query.appBar_ButtonDataBase();
    dataBaseButton.onclick = async () => {
        routerTargetElements().devModeItems.forEach((child) => {
            if ((child as HTMLElement).style.display === "flex") {
                (child as HTMLElement).style.display = "none";
            } else {
                (child as HTMLElement).style.display = "flex";
            }
        });
    };
    dataBaseButton.style.display = "inline-flex";
    cleanup.push(() => {
        dataBaseButton.style.display = "none";
        dataBaseButton.onclick = null;
    });
}

function render() {
    const itemProps: create.GistItemProps[] = [
        {
            title: "Alarm Listen",
            storeKey: "alert-lists",
        },
        {
            title: "Blech Listen",
            storeKey: "metal-sheets",
        },
        {
            title: "VIS",
            storeKey: "vis",
        },
        {
            title: "VIS: Data",
            storeKey: "vis-data",
        },
        {
            title: "Spezial",
            storeKey: "special",
        },
    ];

    const rt = routerTargetElements();
    for (const itemProp of itemProps) {
        const item = create.gistItem(itemProp);
        cleanup.push(item.destroy);
        rt.gistIDs.appendChild(item.element);
    }
}

function routerTargetElements() {
    const rt = query.routerTarget();

    return {
        devModeItems: rt.querySelectorAll(`.dev-mode`),
        gistIDs: rt.querySelector<HTMLUListElement>(`.gist-ids`)!,
    };
}
