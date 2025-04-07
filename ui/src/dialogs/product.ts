import * as types from "../types";

function init(product?: types.Product | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        inputs: {
            lotto: HTMLInputElement;
            name: HTMLInputElement;
            format: HTMLInputElement;
            thickness: HTMLInputElement;
            stamp: HTMLInputElement;
        };
    },
    { open: () => Promise<types.Product | null> }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="product"]`)!;
    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        inputs: {
            lotto: root.querySelector<HTMLInputElement>(`.product-input#product_Lotto`)!,
            name: root.querySelector<HTMLInputElement>(`.product-input#product_Name`)!,
            format: root.querySelector<HTMLInputElement>(`.product-input#product_Format`)!,
            thickness: root.querySelector<HTMLInputElement>(`.product-input#product_Thickness`)!,
            stamp: root.querySelector<HTMLInputElement>(`.product-input#product_Stamp`)!,
        },
    };

    const open: () => Promise<types.Product | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: types.Product | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (e) => {
                e.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                result = {
                    lotto: query.inputs.lotto.value,
                    name: query.inputs.name.value,
                    format: query.inputs.format.value,
                    thickness: parseFloat(query.inputs.thickness.value || "0"),
                    stamp: query.inputs.stamp.value,
                };
            };

            if (!!product) {
                query.inputs.lotto.value = product.lotto;
                query.inputs.name.value = product.name;
                query.inputs.format.value = product.format;
                query.inputs.thickness.value = product.thickness.toString();
                query.inputs.stamp.value = product.stamp;
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
