import * as types from "../types";

function init(vis?: types.Vis | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        inputs: {
            title: HTMLInputElement;
        };
    },
    { open: () => Promise<types.Vis | null> }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="vis"]`)!;
    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        inputs: {
            title: root.querySelector<HTMLInputElement>(`input[type="text"]`)!,
        },
    };

    const open: () => Promise<types.Vis | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: types.Vis | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (e) => {
                e.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                const date = new Date();

                result = {
                    date: date.getTime(),
                    title: query.inputs.title.value || generateDefaultTitle(date),
                    data: vis?.data || [],
                };
            };

            if (!!vis) {
                // Set title input (default: YYYY-MM-DD)
                query.inputs.title.value = vis.title || generateDefaultTitle(new Date());
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

function generateDefaultTitle(date: Date) {
    const month = (date.getMonth() + 1).toString().padStart(2, "0");
    const day = date.getDate().toString().padStart(2, "0");
    return `${date.getFullYear()}-${month}-${day}`;
}

export default init;
