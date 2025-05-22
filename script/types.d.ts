export declare global {
    interface Window {
        ui: typeof import("ui");
    }

    interface Component<T = any> {
        render(target: HTMLElement): Promise<void>;
        update(data: T): Promise<void>;
        destroy(): Promise<void>;
    }

    interface MetalSheetTableCell<T extends string | number | SacmiThickness> {
        valueType: string;
        value: T;
    }

    interface MetalSheetTable {
        dataSearch: string;
        head: string[];
        body: Cell[][];
        hiddenCells: number[];
    }

    interface SacmiThickness {
        current: number;
        max: number;
    }
}
