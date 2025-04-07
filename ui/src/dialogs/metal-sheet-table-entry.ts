import * as globals from "../globals";
import * as types from "../types";

function init(data?: string[] | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        labels: NodeListOf<HTMLLabelElement>;
        inputs: NodeListOf<HTMLInputElement>;
    },
    { open: () => Promise<string[] | null> }
> {
    const root = document.querySelector<HTMLDialogElement>(
        `dialog[name="metal-sheet-table-entry"]`,
    )!;

    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        labels: root.querySelectorAll<HTMLLabelElement>(`label[for]`)!,
        inputs: root.querySelectorAll<HTMLInputElement>(`input[type="text"]`)!,
    };

    const open: () => Promise<string[] | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: string[] | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (event) => {
                event.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                // Get the values from the dialog form inputs
                result = [];
                query.inputs.forEach((input) => {
                    result!.push(input.value);
                });
            };

            query.labels.forEach((label, index) => {
                label.innerText = globals.metalSheetSlots[index] || "";
            });

            if (!!data) {
                data.forEach((value, index) => {
                    query.inputs[index].value = value;
                });
            }

            root.showModal();
        });
    };

    return {
        element: root,
        query,
        utils: {
            open,
        },
        destroy() {},
    };
}

export default init;
