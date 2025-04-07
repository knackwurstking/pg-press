import * as globals from "../globals";
import * as types from "../types";

const html = String.raw;

function init(entry?: types.SpecialFlakesEntry | null): types.Component<
    HTMLDialogElement,
    {
        form: HTMLFormElement;
        inputContainer: HTMLElement;
    },
    { open: () => Promise<types.SpecialFlakesEntry | null> }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="special-flakes-entry"]`)!;
    const cancel = root.querySelector<HTMLButtonElement>(`button.cancel`)!;

    const query = {
        form: root.querySelector<HTMLFormElement>(`form`)!,
        inputContainer: root.querySelector<HTMLElement>(`.input-container`)!,
    };

    const open: () => Promise<types.SpecialFlakesEntry | null> = () => {
        return new Promise((resolve, _reject) => {
            let result: types.SpecialFlakesEntry | null = null;

            root.onclose = () => {
                resolve(result);
            };

            cancel.onclick = (e) => {
                e.preventDefault();
                root.close();
            };

            query.form.onsubmit = () => {
                // Get the values from the dialog form inputs
                const inputElements = query.inputContainer.querySelectorAll(`input`);
                const select = query.inputContainer.querySelector<HTMLSelectElement>("select")!;

                result = {
                    press:
                        entry?.press ||
                        (select.options[select.selectedIndex]
                            .value as types.SpecialFlakes_PressSlot),

                    compatatore: parseInt(inputElements[1].value, 10),

                    primary: {
                        ...(entry?.primary || {}),
                        value: parseFloat(inputElements[0].value || "0", 10),
                        percent: 0,
                    },

                    secondary: [...query.inputContainer.querySelectorAll(`.tower-slot`)]
                        .map((el, index) => {
                            const [inputPercent, inputSpeed] = el.querySelectorAll("input");
                            const percent = parseFloat(inputPercent.value);
                            const speed = parseInt(inputSpeed.value, 10);
                            return {
                                percent,
                                value: speed,
                                slot: globals.flakesTowerSlots[index],
                            };
                        })
                        .filter((e) => !!e.value),
                };

                result.primary.percent = eval(
                    "100 -" + result.secondary.map((e) => `${e.percent}`).join(" - "),
                );
            };

            query.inputContainer.innerHTML = html``;

            const firstRow = document.createElement("div");
            query.inputContainer.appendChild(firstRow);

            {
                firstRow.className = "ui-flex-grid-row";
                firstRow.style.setProperty("--align", "center");
                firstRow.style.setProperty("--justify", "space-evenly");
                firstRow.innerHTML = html`
                    <div class="ui-flex-grid ui-flex-grid-item" style="--gap: 0; --flex: 0;">
                        <label>Main</label>
                        <input
                            class="input"
                            id="specialFlakesEntry_Main"
                            style="width: 12ch; text-align: center;"
                            type="number"
                            value="${entry?.primary.value || ""}"
                            min="0"
                            step="0.1"
                        />
                    </div>

                    <div class="ui-flex-grid ui-flex-grid-item" style="--gap: 0; --flex: 0;">
                        <label>C1</label>
                        <input
                            class="input"
                            id="specialFlakesEntry_C1"
                            style="width: 6ch; text-align: center;"
                            type="number"
                            value="${entry?.compatatore || 25}"
                            min="0"
                            step="1"
                        />
                    </div>

                    <div class="ui-flex-grid ui-flex-grid-item" style="--gap: 0; --flex: 0;">
                        <label>&nbsp;</label>
                        <select>
                            <option value="P0">Presse 0</option>
                            <option value="P4">Presse 4</option>
                            <option value="P5">Presse 5</option>
                        </select>
                    </div>
                `;

                firstRow.querySelector<HTMLSelectElement>(`select`)!.selectedIndex =
                    globals.flakesPressSlots.indexOf(entry?.press || globals.flakesPressSlots[0]);
            }

            globals.flakesTowerSlots.forEach((slot) => {
                const { percent, value } = entry?.secondary.find((e) => e.slot === slot) || {
                    percent: 0,
                    value: 0,
                };

                const item = document.createElement("div");
                query.inputContainer.appendChild(item);

                item.className = "tower-slot ui-flex-grid-item";

                item.style.setProperty("--flex", "0");
                item.style.width = "100%";

                item.innerHTML = html`
                    <h5>${slot}</h5>

                    <div
                        class="ui-flex-grid-row"
                        style="--align: center; --justify: space-between; width: 100%"
                    >
                        <div class="ui-flex-grid-item" style="--flex: 1;">
                            <label>Prozent</label>
                            <input
                                class="input"
                                id="specialFlakesEntry_${slot}-percent"
                                style="text-align: center; width: 100%;"
                                type="number"
                                value="${percent || ""}"
                                min="0"
                                step="0.1"
                            />
                        </div>

                        <div class="ui-flex-grid-item" style="--flex: 1;">
                            <label>Geschwindigkeit</label>
                            <input
                                class="input"
                                id="specialFlakesEntry_${slot}-speed"
                                style="text-align: center; width: 100%;"
                                type="number"
                                value="${value || ""}"
                                min="0"
                                step="1"
                            />
                        </div>
                    </div>
                `;
            });

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
