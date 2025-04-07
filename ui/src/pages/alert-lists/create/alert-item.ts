import * as ui from "ui";

import * as types from "../../../types";

const html = String.raw;

export interface AlertItemProps {
    alert: types.Alert;
    enableRouting?: {
        alertIndex: number;
        listKey: string;
    };
}

export function alertItem(props: AlertItemProps): types.Component<HTMLLIElement> {
    const el = document.createElement("li");

    el.className = "alert-item ui-flex nowrap align-center justify-between";

    el.style.width = "100%";
    el.style.padding = "var(--ui-spacing)";
    el.style.borderBottom = "var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)";

    const onClickHandler = () => {
        ui.router.hash.goTo(
            {
                listKey: props.enableRouting!.listKey,
                index: props.enableRouting!.alertIndex.toString(),
            },
            "alert",
        );
    };

    if (!!props.enableRouting) {
        el.role = "button";
        el.style.cursor = "pointer";

        el.addEventListener("click", onClickHandler);
    }

    el.innerHTML = html`
        <p>${props.alert.alert}</p>
        <p
            style="${ui.styles({
                color: "var(--ui-primary)",
                textWrap: "nowrap",
                marginLeft: "var(--ui-spacing)",
            } as CSSStyleDeclaration)}"
        >
            ${props.alert.from === props.alert.to
                ? props.alert.from
                : `${props.alert.from}..${props.alert.to}`}
        </p>
    `;

    return {
        element: el,
        destroy() {
            if (!!props.enableRouting) {
                el.removeEventListener("click", onClickHandler);
            }
        },
    };
}
