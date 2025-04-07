import * as ui from "ui";

export type CleanUp = () => void;

export interface Alert {
    from: number;
    to: number;
    alert: string;
    desc: string[];
}

export interface AlertList {
    title: string;
    data: Alert[];
}

export interface Bookmarks {
    title: string;
    allowDeletion: boolean;
    data: Product[];
}

export interface Component<
    T extends HTMLElement,
    E = { [key: string]: () => HTMLElement },
    C = { [key: string]: any },
> {
    element: T;
    query?: E;
    utils?: C;
    destroy(): void;
}

export interface Gist {
    id: string;
    revision: number | null;
    token: string;
}

export type DrawerGroups =
    | "alert-lists"
    | "metal-sheets"
    | "vis"
    | "vis-bookmarks"
    | "vis-data"
    | "special";

export interface MetalSheet {
    format: string;
    toolID: string;
    data: {
        /**
         * Use `-1` for none
         */
        press: number;
        table: {
            /**
             * This will hide certain data indexes
             */
            filter?: number[];
            data: string[][];
        };
    };
}

export type PGStore = ui.Store<PGStoreEvents>;

// NOTE:
//  - Kicked: theme.name in v2.0.0
export interface PGStoreEvents {
    "alert-lists": {
        gist: Gist | null;
        lists: AlertList[];
    };

    "metal-sheets": {
        gist: Gist | null;
        lists: MetalSheet[];
    };

    special: {
        gist: Gist | null;
        lists: Special[];
    };

    vis: {
        gist: Gist | null;
        lists: Vis[];
    };

    "vis-data": {
        gist: Gist | null;
        lists: VisData[];
    };

    "vis-bookmarks": {
        gist: Gist | null;
        lists: Bookmarks[];
    };

    runtime: {
        lists: {
            [key: string]: {
                remoteRevision: number | null;
            };
        };
    };
}

export interface Product {
    lotto: string;
    name: string;
    format: string;
    thickness: number;
    stamp: string;
}

export type Special = SpecialFlakes;

export interface SpecialFlakes {
    type: "flakes";
    title: string;
    data: SpecialFlakesEntry[];
}

export interface SpecialFlakesEntry {
    press: SpecialFlakes_PressSlot;
    compatatore: number;
    primary: SpecialFlakes_Consumption;
    secondary: ({ slot: SpecialFlakes_TowerSlot } & SpecialFlakes_Consumption)[];
}

export interface SpecialFlakes_Consumption {
    percent: number;
    value: number;
}

export type SpecialFlakes_TowerSlot = "A" | "C" | "E" | "G" | "I" | "K";
export type SpecialFlakes_PressSlot = "P0" | "P4" | "P5";

export interface Vis {
    /**
     * Needed for vis bookmarks
     * - Updated after each change
     */
    date: number;
    title: string;
    data: Product[];
}

export interface VisData {
    title: string;
    data: VisDataEntry[];
}

export interface VisDataEntry {
    key: string | null;
    value: string;
    lotto: string | null;
    format: string | null;
    thickness: string | null;
    stamp: string | null;
}
