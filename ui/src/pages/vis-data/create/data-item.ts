import * as convert from "../../../convert";
import * as types from "../../../types";

const html = String.raw;

export interface GistItemProps {
    entry: types.VisDataEntry;
    renderTags: boolean;
}

export function dataItem(props: GistItemProps): types.Component<HTMLLIElement> {
    const li = document.createElement("li");

    li.style.borderBottom = "var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)";

    li.innerHTML = html`<div class="ui-flex-grid"></div>`;

    const grid = li.querySelector(".ui-flex-grid")!;

    // Render (optional) key
    if (!!props.entry.key) {
        const div = document.createElement("div");
        div.innerHTML = html`<div class="ui-flex-grid-item"><h4>${props.entry.key}<h4></div>`;
        grid.appendChild(div);
    }

    // Render tags
    if (props.renderTags) {
        const filterGrid = document.createElement("div");
        filterGrid.classList.add("ui-flex-grid");
        filterGrid.style.setProperty("--gap", "0");
        grid.appendChild(filterGrid);
        filterGrid.innerHTML = html`
            <div class="ui-flex-grid-item" style="font-size: 0.9rem"><i>Filter:</i></div>
            <div class="ui-flex-grid-row" style="--gap: var(--ui-spacing);"></div>
        `;

        const gridRow = filterGrid.querySelector(".ui-flex-grid-row")!;

        if (!!props.entry.lotto) {
            gridRow.innerHTML += html`
                <div class="ui-flex-grid-item" style="--flex: 0;">
                    <code>${props.entry.lotto}</code>
                </div>
            `;
        }

        if (!!props.entry.format) {
            gridRow.innerHTML += html`
                <div class="ui-flex-grid-item" style="--flex: 0;">
                    <code>${props.entry.format}</code>
                </div>
            `;
        }

        if (!!props.entry.stamp) {
            gridRow.innerHTML += html`
                <div class="ui-flex-grid-item" style="--flex: 0;">
                    <code>${props.entry.stamp}</code>
                </div>
            `;
        }

        if (!!props.entry.thickness) {
            gridRow.innerHTML += html`
                <div class="ui-flex-grid-item" style="--flex: 0;">
                    <code>${props.entry.thickness}</code><span style="display: none">mm</span>
                </div>
            `;
        }
    }

    // Render value
    const div = document.createElement("div");
    div.innerHTML = html`
        <div class="ui-flex-grid-item" style="width: 100%;">
            <p>${convert.textToHTML(props.entry.value)}</p>
        </div>
    `;
    grid.appendChild(div);

    return {
        element: li,
        destroy() {},
    };
}
