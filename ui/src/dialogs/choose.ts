import * as types from "../types";

function init(
    title: string,
    choices: string[],
): types.Component<
    HTMLDialogElement,
    {
        close: HTMLButtonElement;
        title: HTMLElement;
        choices: HTMLElement;
    },
    {
        open: () => Promise<string | null>;
    }
> {
    const root = document.querySelector<HTMLDialogElement>(`dialog[name="choose"]`)!;

    const query = {
        close: root.querySelector<HTMLButtonElement>(`button.close`)!,
        title: root.querySelector<HTMLElement>(`.title`)!,
        choices: root.querySelector<HTMLElement>(`.choices`)!,
    };

    const open: () => Promise<string | null> = () => {
        return new Promise((resolve, _reject) => {
            let currentChoice: string | null = null;

            query.close.onclick = () => {
                root.close();
            };

            root.onclose = () => {
                resolve(currentChoice);
            };

            query.title.innerText = title;

            // Create dialog button for each choice
            const choicesElement = query.choices;
            choicesElement.innerHTML = "";
            for (const choice of choices) {
                const button = document.createElement("button");

                button.innerText = choice;

                button.setAttribute("variant", "ghost");
                button.setAttribute("color", "secondary");

                button.onclick = () => {
                    currentChoice = button.innerText;
                    root.close();
                };

                choicesElement.appendChild(button);
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
