import * as base from "../base";
import * as types from "../../types";

export class VISData extends base.ListStore<"vis-data"> {
    public key(): keyof base.ListStoreData {
        return "vis-data";
    }

    public listKey(list: types.VisData): string {
        return `${list.title}`;
    }

    public title(): string {
        return "Vis Data";
    }

    public fileName(list: types.VisData): string {
        return `Vis Data - ${super.fileName(list)}`;
    }

    public validate(dataString: string): types.VisData | null {
        const list = super.validate(dataString);
        if (typeof list !== "object") {
            return null;
        }

        if (typeof list.title !== "string" || !Array.isArray(list.data)) {
            return null;
        }

        for (const part of list.data) {
            if (typeof part.key !== "string" && part.key !== null) {
                return null;
            }

            if (typeof part.value !== "string") {
                return null;
            }

            if (typeof part.lotto !== "string" && part.lotto !== null) {
                return null;
            }

            if (typeof part.format !== "string" && part.format !== null) {
                return null;
            }

            if (typeof part.thickness !== "string" && part.thickness !== null) {
                return null;
            }

            if (typeof part.stamp !== "string" && part.stamp !== null) {
                return null;
            }
        }

        return list;
    }
}
