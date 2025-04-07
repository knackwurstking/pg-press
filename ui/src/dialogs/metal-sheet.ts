import * as types from "../types";

function init(metalSheet?: types.MetalSheet | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        inputs: {
            format: HTMLInputElement;
            toolID: HTMLInputElement;
            press: HTMLSelectElement;
        };
    },
    {
        open: () => Promise<types.MetalSheet | null>;
        validate: () => boolean;
    }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="metal-sheet"]`)!;
    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        inputs: {
            format: root.querySelector<HTMLInputElement>(`input#metalSheetDialog_Format`)!,
            toolID: root.querySelector<HTMLInputElement>(`input#metalSheetDialog_ToolID`)!,
            press: root.querySelector<HTMLSelectElement>(`select#metalSheetDialog_Press`)!,
        },
    };

    const open: () => Promise<types.MetalSheet | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: types.MetalSheet | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (e) => {
                e.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                const press = parseInt(
                    (
                        query.inputs.press.children[
                            query.inputs.press.selectedIndex
                        ] as HTMLOptionElement
                    ).value,
                    10,
                );

                result = {
                    format: query.inputs.format.value,
                    toolID: query.inputs.toolID.value,
                    data: {
                        press: press,
                        table: {
                            filter: metalSheet?.data.table.filter || [],
                            data: metalSheet?.data.table.data || [],
                        },
                    },
                };
            };

            if (!!metalSheet) {
                query.inputs.format.value = metalSheet.format;
                query.inputs.toolID.value = metalSheet.toolID;

                query.inputs.press.selectedIndex = 0;
                let index = 0;
                for (const option of query.inputs.press.options) {
                    if (option.value === metalSheet.data.press.toString()) {
                        query.inputs.press.selectedIndex = index;
                        break;
                    }

                    index++;
                }
            }

            root.showModal();
        });
    };

    return {
        element: root,
        query,
        utils: {
            open,
            validate() {
                let valid = true;

                if (!query.inputs.format.value) {
                    query.inputs.format.ariaInvalid = "";
                    valid = false;
                } else {
                    query.inputs.format.ariaInvalid = null;
                }

                if (!query.inputs.toolID.value) {
                    query.inputs.toolID.ariaInvalid = "";
                    valid = false;
                } else {
                    query.inputs.toolID.ariaInvalid = null;
                }

                return valid;
            },
        },
        destroy() {},
    };
}

export default init;
