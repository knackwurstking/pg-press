import * as types from "../../../types";

const html = String.raw;

export interface ProductItemProps {
    product: types.Product;
    enableRouting?: {
        dataIndex: number;
    };
}

export function productItem(props: ProductItemProps): types.Component<HTMLLIElement> {
    const el = document.createElement("li");

    el.className = "product-item ui-flex nowrap align-center justify-between";

    el.style.width = "100%";
    el.style.padding = "var(--ui-spacing)";
    el.style.borderBottom = "var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)";

    if (!!props.enableRouting) {
        el.role = "button";
        el.style.cursor = "pointer";

        el.setAttribute("data-index", props.enableRouting!.dataIndex.toString());
    }

    props.product.lotto;
    props.product.name;
    el.innerHTML = html`
        <div class="ui-flex-grid" style="--gap: 0.5rem; --mono: 1;">
            <div
                class="ui-flex-grid-row"
                style="--gap: 1rem; --justify: space-between; --wrap: wrap;"
            >
                <div class="lotto" style="--wght: 700; color: var(--ui-primary);">
                    ${props.product.lotto}
                </div>
                <div class="name" style="--wght: 350; --slnt: -4;">${props.product.name}</div>
            </div>

            <div
                class="ui-flex-grid-row"
                style="--gap: 1rem; --justify: space-between; --wrap: wrap;"
            >
                <div class="ui-flex-grid-row" style="--justify: flex-start; width: fit-content">
                    <div class="format">${props.product.format}</div>
                    <div class="thickness">${props.product.thickness}mm</div>
                </div>

                <div class="stamp" style="text-wrap: nowrap;">${props.product.stamp}</div>
            </div>
        </div>
    `;

    return {
        element: el,
        destroy() {},
    };
}
