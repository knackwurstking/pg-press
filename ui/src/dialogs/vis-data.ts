import * as types from "../types";

function init(visData?: types.VisData | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        inputs: {
            title: HTMLInputElement;
        };
    },
    {
        open: () => Promise<types.VisData | null>;
        validate: () => boolean;
    }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="vis-data"]`)!;
    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        inputs: {
            title: root.querySelector<HTMLInputElement>(`input[type="text"]`)!,
        },
    };

    const open: () => Promise<types.VisData | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: types.VisData | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (e) => {
                e.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                result = {
                    title: query.inputs.title.value,
                    data: visData?.data || [],
                };
            };

            if (!!visData) {
                query.inputs.title.value = visData.title;
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

                if (!query.inputs.title.value) {
                    query.inputs.title.ariaInvalid = "";
                    valid = false;
                } else {
                    query.inputs.title.ariaInvalid = null;
                }

                return valid;
            },
        },
        destroy() {},
    };
}

export default init;
