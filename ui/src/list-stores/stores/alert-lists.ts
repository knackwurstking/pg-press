import * as types from "../../types";
import * as base from "../base";

export class AlertLists extends base.ListStore<"alert-lists"> {
    public key(): keyof base.ListStoreData {
        return "alert-lists";
    }

    public listKey(list: types.AlertList): string {
        return list.title;
    }

    public title(): string {
        return "Alarm Listen";
    }

    public fileName(list: types.AlertList): string {
        return `Alarm Liste - ${super.fileName(list)}`;
    }

    public validate(dataString: any): types.AlertList | null {
        const list = super.validate(dataString);
        if (typeof list !== "object") {
            return null;
        }

        if (typeof list?.title !== "string") {
            return null;
        }

        if (!("data" in list)) {
            return null;
        }

        if (!Array.isArray(list.data)) {
            return null;
        }

        for (const part of list.data) {
            if (!("from" in part && "to" in part && "alert" in part && "desc" in part)) {
                return null;
            }

            if (
                typeof part.from !== "number" ||
                typeof part.to !== "number" ||
                typeof part.alert !== "string"
            ) {
                return null;
            }

            if (typeof part.desc === "string") {
                part.desc = part.desc.split("\n");
            }

            if (!Array.isArray(part.desc)) {
                return null;
            }

            if ((part.desc as any[]).filter((line) => typeof line !== "string").length > 0) {
                return null;
            }
        }

        return list;
    }
}
