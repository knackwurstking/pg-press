import * as base from "../base";
import * as types from "../../types";

export class Special extends base.ListStore<"special"> {
    public key(): keyof base.ListStoreData {
        return "special";
    }

    public listKey(list: types.Special): string {
        return `${list.title}`;
    }

    public title(): string {
        return "Special";
    }

    public fileName(list: types.Special): string {
        return `Special - ${super.fileName(list)}`;
    }

    public validate(dataString: string): types.Special | null {
        const list = super.validate(dataString);
        if (typeof list !== "object") {
            return null;
        }

        if (typeof list.title !== "string") {
            return null;
        }

        switch (list.type) {
            case "flakes":
                if (validateFlakesData(list.data)) {
                    return list;
                }
                break;
        }

        return null;
    }
}

function validateFlakesData(data: any): boolean {
    if (!Array.isArray(data)) {
        return false;
    }

    for (const part of data) {
        if (typeof part !== "object") {
            return false;
        }

        if (!["P0", "P4", "P5"].includes(part.press)) {
            return false;
        }

        if (typeof part.compatatore !== "number") {
            return false;
        }

        if (typeof part.primary !== "object") {
            return false;
        }

        if (typeof part.primary.percent !== "number") {
            return false;
        }

        if (typeof part.primary.value !== "number") {
            return false;
        }

        if (!Array.isArray(part.secondary)) {
            return false;
        }

        for (const consumption of part.secondary) {
            if (!["A", "C", "E", "G", "I", "K"].includes(consumption.slot)) {
                return false;
            }

            if (typeof consumption.percent !== "number") {
                return false;
            }

            if (typeof consumption.value !== "number") {
                return false;
            }
        }
    }

    return true;
}
