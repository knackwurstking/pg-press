import * as types from "../types";

function init(product?: types.VisDataEntry | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        inputs: {
            key: HTMLInputElement;
            value: HTMLInputElement;
            lotto: HTMLInputElement;
            format: HTMLInputElement;
            thickness: HTMLInputElement;
            stamp: HTMLInputElement;
        };
    },
    {
        open: () => Promise<types.VisDataEntry | null>;
        validate: () => boolean;
    }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="vis-data-entry"]`)!;
    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        inputs: {
            key: root.querySelector<HTMLInputElement>(`.vis-data-entry-input#visDataEntry_Key`)!,
            value: root.querySelector<HTMLInputElement>(
                `.vis-data-entry-input#visDataEntry_Value`,
            )!,
            lotto: root.querySelector<HTMLInputElement>(
                `.vis-data-entry-input#visDataEntry_Lotto`,
            )!,
            format: root.querySelector<HTMLInputElement>(
                `.vis-data-entry-input#visDataEntry_Format`,
            )!,
            thickness: root.querySelector<HTMLInputElement>(
                `.vis-data-entry-input#visDataEntry_Thickness`,
            )!,
            stamp: root.querySelector<HTMLInputElement>(
                `.vis-data-entry-input#visDataEntry_Stamp`,
            )!,
        },
    };

    const open: () => Promise<types.VisDataEntry | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: types.VisDataEntry | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (e) => {
                e.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                result = {
                    key: query.inputs.key.value || null,
                    value: query.inputs.value.value,
                    lotto: query.inputs.lotto.value || null,
                    format: query.inputs.format.value || null,
                    thickness: query.inputs.thickness.value || null,
                    stamp: query.inputs.stamp.value || null,
                };
            };

            if (!!product) {
                query.inputs.key.value = product.key || "";
                query.inputs.value.value = product.value || "";
                query.inputs.lotto.value = product.lotto || "";
                query.inputs.format.value = product.format || "";
                query.inputs.thickness.value = product.thickness || "";
                query.inputs.stamp.value = product.stamp || "";
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
                if (!query.inputs.value.value) query.inputs.value.ariaInvalid = "";
                else query.inputs.value.ariaInvalid = null;

                return query.inputs.value.ariaInvalid === null;
            },
        },
        destroy() {},
    };
}

export default init;
