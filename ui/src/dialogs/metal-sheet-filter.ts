import * as globals from "../globals";
import * as types from "../types";

const html = String.raw;

function init(filter?: number[] | null): types.Component<
    HTMLDialogElement,
    {
        filters: HTMLUListElement;
    },
    {
        open: () => Promise<number[] | null>;
    }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="metal-sheet-filter"]`)!;

    const query = {
        filters: root.querySelector<HTMLUListElement>(`ul.filters`)!,
    };

    const open: () => Promise<number[] | null> = () => {
        return new Promise((resolve, _reject) => {
            root.onclose = () => {
                const newFilter: number[] = [];

                query.filters
                    .querySelectorAll<HTMLInputElement>(`input[type="checkbox"]`)
                    .forEach((checkbox) => {
                        if (checkbox.checked) return;

                        const indexToHide = parseInt(checkbox.getAttribute("data-index")!, 10);
                        newFilter!.push(indexToHide);
                    });

                resolve(newFilter || []);
            };

            query.filters.innerHTML = "";

            globals.metalSheetSlots.forEach((slot, index) => {
                const li = document.createElement("li");
                li.className = "filter";
                li.innerHTML = html`
                    <label>
                        <input type="checkbox" data-index="${index}" checked />
                        <span class="name"></span>
                    </label>
                `;

                li.querySelector<HTMLElement>(`.name`)!.innerText = slot;
                li.querySelector<HTMLInputElement>(`input[type="checkbox"]`)!.checked =
                    !filter?.includes(index);

                query.filters.appendChild(li);
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
