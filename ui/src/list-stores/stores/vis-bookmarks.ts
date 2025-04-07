import * as base from "../base";
import * as types from "../../types";

export class VISBookmarks extends base.ListStore<"vis-bookmarks"> {
    public key(): keyof base.ListStoreData {
        return "vis-bookmarks";
    }

    public listKey(list: types.Bookmarks): string {
        return `${list.title}`;
    }

    public title(): string {
        return "Vis Bookmarks";
    }

    public fileName(list: types.Bookmarks): string {
        return `Vis Bookmarks - ${super.fileName(list)}`;
    }
}
