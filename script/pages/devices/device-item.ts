type onClickPowerButton = (item: HTMLElement) => void | Promise<void>;

export function create(
    device: Device,
    onClickPowerButton: onClickPowerButton,
): HTMLElement {
    const t = document.querySelector<HTMLTemplateElement>(
        `template[name="device-list-item"]`,
    )!;

    const item = (
        t.content.cloneNode(true) as HTMLElement
    ).querySelector<HTMLElement>(".device-list-item")!;

    return update(item, device, onClickPowerButton);
}

export function update(
    item: HTMLElement,
    device: Device,
    onClickPowerButton: onClickPowerButton,
): HTMLElement {
    item.setAttribute("data-addr", device.server.addr);

    const title = item.querySelector<HTMLElement>(`.title`)!;
    title.innerHTML = device.server.name || device.server.addr;

    const editButton = item.querySelector<HTMLElement>(`button.edit`)!;
    editButton.setAttribute("data-addr", device.server.addr);

    const powerButton =
        item.querySelector<HTMLButtonElement>(`button.power-button`)!;

    powerButton.onclick = async (ev) => {
        const target = ev.currentTarget as HTMLButtonElement;

        if (target.getAttribute("data-state") === "processing") {
            return;
        }

        target.setAttribute("data-state", "processing");

        await onClickPowerButton(item);
    };

    const background = powerButton.querySelector<HTMLElement>(`.background`)!;

    background.style.backgroundColor = `rgb(${[...(device.color || []), 0, 0, 0].slice(0, 3).join(", ")})`;

    if (Math.max(...(device.color || []), 0, 0, 0)) {
        powerButton.setAttribute("data-state", "on");
    } else {
        powerButton.setAttribute("data-state", "off");
    }

    return item;
}
