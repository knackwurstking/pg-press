import * as types from "./types";

// Router

export function routerTarget(): HTMLElement {
    return document.querySelector(`#routerTarget`)!;
}

// App Bar Elements

export function appBar_ButtonOpenDrawer(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonOpenDrawer`)!;
}

export function appBar_ButtonBack(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonBack`)!;
}

export function appBar_Title(): HTMLHeadingElement {
    return document.querySelector(`#appBarTitle`)!;
}

export function appBar_ButtonListEdit(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonListEdit`)!;
}

export function appBar_ButtonAdd(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonAdd`)!;
}

export function appBar_ButtonBookmarks(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonBookmarks`)!;
}

export function appBar_ButtonDataBase(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonDataBase`)!;
}

export function appBar_ButtonPrinter(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonPrinter`)!;
}

export function appBar_ButtonFilter(): HTMLButtonElement {
    return document.querySelector(`#appBarButtonFilter`)!;
}

// Drawer Elements

export function drawer(): HTMLElement {
    return document.querySelector(`aside.ui-drawer`)!;
}

export function drawerGistIDsButton(): HTMLButtonElement {
    return drawer().querySelector(`button.gist-ids`)!;
}

export function drawerGroup(name: types.DrawerGroups): {
    root: HTMLDetailsElement;
    actions: {
        download: HTMLButtonElement | null;
        add: HTMLButtonElement | null;
        importFromFile: HTMLButtonElement | null;
    };
    items: HTMLUListElement;
} {
    const group = drawer().querySelector(`.group[name="${name}"]`)!;
    return {
        root: group as HTMLDetailsElement,
        actions: {
            download: group.querySelector(`button.download`) || null,
            add: group.querySelector(`button.add`) || null,
            importFromFile: group.querySelector(`button.import-from-file`) || null,
        },
        items: group.querySelector(`ul.items`)!,
    };
}

export function drawerBackdrop(): HTMLElement {
    return document.querySelector(`div.ui-drawer-backdrop`)!;
}
