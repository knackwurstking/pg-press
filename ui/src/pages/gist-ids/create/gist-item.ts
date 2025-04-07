import * as gist from "../../../gist";
import * as globals from "../../../globals";
import * as types from "../../../types";

const html = String.raw;

export interface GistItemProps {
    title: string;
    storeKey: types.DrawerGroups;
}

export function gistItem(props: GistItemProps): types.Component<HTMLLIElement> {
    const el = document.createElement("li");

    el.className = "gist-item ui-flex-grid-item ui-border";
    el.style.width = "100%";

    el.innerHTML = html`
        <div class="ui-flex-grid-row" style="--justify: space-between; --align: center;">
            <h3>${props.title}</h3>
            <div
                class="ui-flex-grid"
                style="--align: flex-end; --gap: 0; --mono: 1; font-size: 0.85rem; width: fit-content;"
            >
                <span>
                    <span>Remote Rev.: </span>
                    <span id="gistID_RemoteRevision_${props.storeKey}">?</span>
                </span>

                <span>
                    <span>Local Rev.: </span>
                    <span
                        id="gistID_LocalRevision_${props.storeKey}"
                        style="--wght: 450; color: var(--ui-text)"
                    >
                        ${globals.store.get(props.storeKey)!.gist?.revision || "?"}
                    </span>
                </span>
            </div>
        </div>

        <div
            class="gist-id-container ui-flex-grid-row"
            style="--align: flex-end; --justify: space-between; width: 100%;"
        >
            <div class="ui-flex-grid-item">
                <label for="gistID_${props.storeKey}">Gist ID</label>
                <input
                    id="gistID_${props.storeKey}"
                    style="width: 100%"
                    type="text"
                    placeholder="Gist ID von Telegram hier einfügen"
                    value="${globals.store.get(props.storeKey)!.gist?.id || ""}"
                />
            </div>
            <div class="ui-flex-grid-item" style="--flex: 0;">
                <button class="update" variant="ghost" color="primary" icon>
                    <i class="bi bi-cloud-download"></i>
                </button>
            </div>
        </div>

        <div
            class="dev-mode ui-flex-grid-row"
            style="--align: flex-end; --justify: space-between; width: 100%; display: none"
        >
            <div class="ui-flex-grid-item">
                <label for="apiToken_${props.storeKey}">API Token</label>
                <input
                    id="apiToken_${props.storeKey}"
                    style="width: 100%"
                    type="text"
                    placeholder="API Token hier einfügen"
                    value="${globals.store.get(props.storeKey)!.gist?.token || ""}"
                />
            </div>
            <div class="ui-flex-grid-item" style="--flex: 0;">
                <button class="push" variant="ghost" color="primary" icon>
                    <i class="bi bi-cloud-upload"></i>
                </button>
            </div>
        </div>
    `;

    initNormalMode(el, props);
    initDevMode(el, props);

    return {
        element: el,
        destroy() {},
    };
}

async function initNormalMode(item: HTMLLIElement, props: GistItemProps) {
    const elements = queryElements(item, props);
    elements.inputGistID.onchange = async () => elements.updateButton.click();

    let inProgress: boolean = false;
    elements.updateButton.onclick = async () => {
        if (inProgress) {
            return;
        }

        inProgress = true;
        elements.updateButton.classList.add("active");
        const cleanUp = () => {
            setTimeout(() => {
                inProgress = false;
                elements.updateButton.classList.remove("active");
            });
        };

        const gistID = elements.inputGistID.value;

        if (!gistID) {
            globals.store.update(props.storeKey, (data) => {
                data.gist = null;
                return data;
            });

            return cleanUp();
        }

        globals.store.update(props.storeKey, (data) => {
            data.gist = {
                id: gistID,
                revision: null,
                token: elements.inputAPIToken.value,
            };
            return data;
        });

        try {
            const data = await gist.pull(props.storeKey, gistID);
            globals.store.set(props.storeKey, data);

            elements.remoteRevSpan.innerText =
                elements.localRevSpan.innerText = `${data.gist?.revision || "?"}`;

            elements.localRevSpan.style.color = "green";
            elements.updateButton.ariaInvalid = null;
        } catch (err) {
            const message = `Pull from gist failed for "${props.storeKey}" ("${gistID}"): ${err}`;
            console.error(message);
            alert(message);
            elements.updateButton.ariaInvalid = "";
        }

        return cleanUp();
    };

    if (!!elements.inputGistID.value) {
        setTimeout(async () => {
            let remoteRev =
                globals.store.get("runtime")!.lists[elements.inputGistID.value]?.remoteRevision ||
                null;
            if (!remoteRev) {
                try {
                    remoteRev = await gist.getRevision(elements.inputGistID.value);

                    globals.store.update("runtime", (data) => {
                        data.lists[elements.inputGistID.value] = {
                            remoteRevision: remoteRev,
                        };
                        return data;
                    });
                } catch (err) {
                    elements.inputGistID.ariaInvalid = "";

                    console.error(
                        `Update failed for "${props.storeKey}" ("${elements.inputGistID.value}"): ${err}`,
                    );
                }
            }

            elements.remoteRevSpan.innerText = `${remoteRev || "?"}`;
            elements.inputGistID.ariaInvalid = null;

            if (remoteRev === null) {
                elements.localRevSpan.style.color = "var(--ui-text)";
            } else if (remoteRev !== globals.store.get(props.storeKey)!.gist?.revision || null) {
                elements.localRevSpan.style.color = "red";
            } else {
                elements.localRevSpan.style.color = "green";
            }
        });
    }
}

async function initDevMode(item: HTMLLIElement, props: GistItemProps) {
    const elements = queryElements(item, props);

    elements.inputAPIToken.oninput = () => {
        if (!!elements.inputAPIToken.value) {
            elements.pushButton.removeAttribute("disabled");
        } else {
            elements.pushButton.setAttribute("disabled", "");
        }

        globals.store.update(props.storeKey, (data) => {
            if (!data.gist) {
                data.gist = {
                    id: "",
                    revision: null,
                    token: elements.inputAPIToken.value,
                };
            } else {
                data.gist.token = elements.inputAPIToken.value;
            }

            return data;
        });
    };
    elements.inputAPIToken.oninput(new Event("input"));

    let pushInProgress = false;
    elements.pushButton.onclick = async () => {
        if (!elements.inputAPIToken.value) {
            return;
        }

        if (pushInProgress) {
            return;
        }

        pushInProgress = true;
        elements.pushButton.classList.add("active");
        const cleanUp = () => {
            setTimeout(() => {
                pushInProgress = false;
                elements.pushButton.classList.remove("active");
            });
        };

        const gistID = elements.inputGistID.value;
        const apiToken = elements.inputAPIToken.value;

        try {
            const data = await gist.push(props.storeKey, apiToken, gistID);
            globals.store.set(props.storeKey, data);

            elements.remoteRevSpan.innerText =
                elements.localRevSpan.innerText = `${data.gist?.revision || "?"}`;

            elements.localRevSpan.style.color = "green";
        } catch (err) {
            const message = `Push to gist failed for "${props.storeKey}" ("${gistID}"): ${err}`;
            console.error(message);
            alert(message);
        }

        return cleanUp();
    };
}

function queryElements(item: HTMLLIElement, props: GistItemProps) {
    return {
        localRevSpan: item.querySelector<HTMLSpanElement>(
            `#gistID_LocalRevision_${props.storeKey}`,
        )!,
        remoteRevSpan: item.querySelector<HTMLSpanElement>(
            `#gistID_RemoteRevision_${props.storeKey}`,
        )!,

        inputGistID: item.querySelector<HTMLInputElement>(`#gistID_${props.storeKey}`)!,
        updateButton: item.querySelector<HTMLButtonElement>(`button.update`)!,

        inputAPIToken: item.querySelector<HTMLInputElement>(`#apiToken_${props.storeKey}`)!,
        pushButton: item.querySelector<HTMLButtonElement>(`button.push`)!,
    };
}
