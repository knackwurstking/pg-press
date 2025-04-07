import * as base from "./base";
import * as stores from "./stores";

export function get<T extends keyof base.ListStoreData>(key: T): base.ListStore<T> {
    switch (key) {
        case "alert-lists":
            return new stores.AlertLists() as base.ListStore<T>;

        case "metal-sheets":
            return new stores.MetalSheets() as base.ListStore<T>;

        case "vis":
            return new stores.VIS() as base.ListStore<T>;

        case "vis-bookmarks":
            return new stores.VISBookmarks() as base.ListStore<T>;

        case "vis-data":
            return new stores.VISData() as base.ListStore<T>;

        case "special":
            return new stores.Special() as base.ListStore<T>;

        default:
            throw new Error(`unknown "${key}"`);
    }
}
