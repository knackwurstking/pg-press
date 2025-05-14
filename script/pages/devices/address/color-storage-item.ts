interface Options {
    device?: Device;
    onClick?: (color: Color) => void | Promise<void>;
    onChange?: (color: Color) => void | Promise<void>;
    enableDelete?: boolean;
    onDelete?: (color: Color) => void | Promise<void>;
}

export const colorSeparator = ",";

export function splitDataColor(v: string): Color {
    return v.split(colorSeparator).map((c) => parseInt(c, 10));
}

export function create(
    index: number,
    color: Color,
    options: Options,
): HTMLElement {
    const t = document.querySelector<HTMLTemplateElement>(
        `template[name="color-storage-item"]`,
    )!;

    const item = (
        t.content.cloneNode(true) as HTMLElement
    ).querySelector<HTMLElement>(`*`)!;

    return update(item, index, color, options);
}

export function update(
    item: HTMLElement,
    index: number,
    color: Color,
    options: Options,
): HTMLElement {
    if (color.length < 3) color = [...color, 0, 0, 0];
    color = color.slice(0, 3);
    item.style.color = `rgb(${color.join(", ")})`;
    item.setAttribute("data-color", `${color.join(colorSeparator)}`);

    if (options?.onClick) {
        item.onclick = () => {
            // @ts-expect-error
            options.onClick(color);
        };
    } else item.onclick = null;

    const input = item.querySelector<HTMLInputElement>(`input`)!;

    input.onchange = (ev) => {
        const target = ev.currentTarget as HTMLInputElement;

        const color = hexToRGB(target.value || "#FFFFFF");
        window.api.setColor(index, color);
        if (options?.device) {
            window.api.setDevicesColor(color, options.device);
            if (options.onChange) options.onChange(color);
        }

        update(item, index, color, options);
    };

    const delBtn = item.querySelector<HTMLButtonElement>(`button.delete`)!;
    if (options?.enableDelete) {
        delBtn.style.display = "";
        delBtn.onclick = async (ev) => {
            ev.stopPropagation();
            if (options.onDelete) options.onDelete(color);
        };
    } else {
        delBtn.style.display = "none";
    }

    return item;
}

export function hexToRGB(value: string) {
    value = (value || "#FFFFFF").slice(1);
    const color = [];

    for (let x = 0; x < value.length; x += 2) {
        color.push(parseInt(value.slice(x, x + 2), 16));
    }

    return color;
}
