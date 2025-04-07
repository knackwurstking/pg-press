import * as types from "../../types";
import * as globals from "../../globals";
import * as listsStore from "../../list-stores";

const html = String.raw;
const ls = listsStore.get("metal-sheets");

export interface MetalSheetItemProps {
    data: types.MetalSheet;
}

export function metalSheetItem(props: MetalSheetItemProps): types.Component<HTMLLIElement> {
    const el = document.createElement("li");

    el.className = "ui-flex justify-between";

    const pressString = props.data.data.press > -1 ? `[P${props.data.data.press}]` : "";

    el.innerHTML = html`
        <a
            class="ui-flex column align-start justify-center"
            style="width: 100%; height: 100%;"
            href="#metal-sheets?listKey=${ls.listKey(props.data)}"
        >
            <span>${pressString} ${props.data.format} ${props.data.toolID}</span>
            <span>${props.data.data.table.data.length} Einträge</span>
        </a>

        <button class="delete" variant="ghost" color="destructive">
            <i class="bi bi-trash"></i>
        </button>
    `;

    const onClickDelete = () => {
        globals.store.update("metal-sheets", (data) => {
            const key = ls.listKey(props.data);

            if (confirm(`"${key}" wirklich löschen?`)) {
                data.lists = data.lists.filter((list) => ls.listKey(list) !== key);
            }

            return data;
        });
    };

    const deleteButton = el.querySelector<HTMLButtonElement>(`button.delete`)!;
    deleteButton.addEventListener("click", onClickDelete);

    return {
        element: el,
        destroy() {
            deleteButton.removeEventListener("click", onClickDelete);
        },
    };
}
