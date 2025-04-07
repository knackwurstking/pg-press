import * as globals from "../globals";
import * as listStores from "../list-stores";
import * as types from "../types";

const html = String.raw;

function init(product: types.Product): types.Component<
    HTMLDialogElement,
    {
        close: HTMLButtonElement;
        title: HTMLElement;
        checkboxes: HTMLUListElement;
    },
    { open: () => Promise<null> }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="product-bookmark"]`)!;

    const query = {
        close: root.querySelector<HTMLButtonElement>(`button.close`)!,
        title: root.querySelector<HTMLElement>(`.title`)!,
        checkboxes: root.querySelector<HTMLUListElement>(`.checkboxes`)!,
    };

    const open: () => Promise<null> = () => {
        return new Promise((resolve, _reject) => {
            root.onclose = () => resolve(null);
            query.close.onclick = () => root.close();

            query.title.innerText = product.lotto;

            // For each bookmarks list
            const ls = listStores.get("vis-bookmarks");
            query.checkboxes.innerHTML = "";
            globals.store.get("vis-bookmarks")!.lists.forEach((bookmarks, index) => {
                // Render checklist item
                const li = document.createElement("li");

                li.innerHTML = html`
                    <label>
                        <input type="checkbox" value="${ls.listKey(bookmarks)}" />
                        ${bookmarks.title}
                    </label>
                `;

                const input = li.querySelector<HTMLInputElement>(`input`)!;

                // Handle input
                input.onchange = () => {
                    globals.store.update("vis-bookmarks", (data) => {
                        if (input.checked) {
                            data.lists[index].data.push(product);
                        } else {
                            data.lists[index].data = data.lists[index].data.filter(
                                (p) => p.lotto !== product.lotto && p.name !== product.name,
                            );
                        }

                        return data;
                    });
                };

                query.checkboxes.appendChild(li);

                for (const bookmarkedProduct of bookmarks.data) {
                    if (
                        bookmarkedProduct.lotto === product.lotto &&
                        bookmarkedProduct.name === product.name
                    ) {
                        input.checked = true;
                        break;
                    } else {
                        input.checked = false;
                    }
                }
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
