import * as types from "../types";
import * as globals from "../globals";

export interface ListStoreData {
    "alert-lists": types.AlertList;
    "metal-sheets": types.MetalSheet;
    vis: types.Vis;
    "vis-data": types.VisData;
    "vis-bookmarks": types.Bookmarks;
    special: types.Special;
}

export class ListStore<T extends keyof ListStoreData> {
    public key(): keyof ListStoreData {
        return "" as keyof ListStoreData;
    }

    public listKey(list: ListStoreData[T]): string {
        if ("title" in list) {
            return list.title as string;
        }

        return "unknown";
    }

    public title(): string {
        return "";
    }

    public fileName(list: ListStoreData[T]): string {
        return `${this.listKey(list)}.json`;
    }

    public zipFileName(): string {
        return `${this.title()} - ${new Date().getTime()}.zip`;
    }

    public validate(dataString: string): any | null {
        try {
            return JSON.parse(dataString);
        } catch {
            return null;
        }
    }

    public sort(data: ListStoreData[T][]): ListStoreData[T][] {
        const result: ListStoreData[T][] = [];

        const keys = data.map((list) => `${this.listKey(list)}`).sort();

        for (const key of keys) {
            const keyData = data.find((list) => `${this.listKey(list)}` === key);
            if (keyData !== undefined) {
                result.push(keyData);
            }
        }

        return result;
    }

    /**
     * @throws exists error
     */
    public replaceInStore(newList: ListStoreData[T], oldList?: ListStoreData[T]) {
        const newListKey = this.listKey(newList);
        const oldListKey = this.listKey(oldList || newList);

        if (oldListKey !== newListKey) {
            for (const list of globals.store.get(this.key() as keyof ListStoreData)!.lists) {
                if (this.listKey(list as any) === newListKey) {
                    throw new Error(`Liste "${newListKey}" existiert bereits!"`);
                }
            }
        }

        globals.store.update(this.key() as keyof ListStoreData, (data) => {
            for (let i = 0; i < data.lists.length; i++) {
                if (this.listKey(data.lists[i] as ListStoreData[T]) === oldListKey) {
                    data.lists[i] = newList;
                }
            }

            data.lists = this.sort(data.lists as ListStoreData[T][]) as any[];

            return data;
        });
    }

    /**
     * @throws exists error
     */
    public addToStore(newLists: ListStoreData[T][]) {
        const storeData = globals.store.get(this.key() as keyof ListStoreData)!;

        // Check
        for (const newList of newLists) {
            const newListKey = this.listKey(newList as ListStoreData[T]);

            const existingData = storeData.lists.find((list) => {
                if (this.listKey(list as ListStoreData[T]) === newListKey) {
                    return true;
                }

                return false;
            });

            if (existingData !== undefined) {
                throw new Error(`Liste "${newListKey}" existiert bereits!"`);
            }
        }

        // Merge
        const mergedData = [
            ...storeData.lists.filter((storeList) => {
                const data = newLists.find(
                    (list) => this.listKey(list) === this.listKey(storeList as ListStoreData[T]),
                );

                return data === undefined;
            }),
            ...newLists,
        ];

        // Store
        globals.store.set(this.key() as keyof ListStoreData, {
            ...storeData,
            lists: this.sort(mergedData as ListStoreData[T][]) as any[],
        });
    }
}
