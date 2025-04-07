import * as types from "../../../types";

const html = String.raw;

export interface AlertDescriptionProps {
    alert: types.Alert;
}

export function alertDescription(props: AlertDescriptionProps): types.Component<HTMLElement> {
    const el = document.createElement("div");

    el.className = "ui-flex column";
    el.style.padding = "var(--ui-spacing)";

    el.innerHTML = html`${props.alert.desc.map((line) => `<p>${line}</p>`).join("")}`;

    return {
        element: el,
        destroy() {},
    };
}
