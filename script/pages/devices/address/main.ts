import * as colorStorageItem from "./color-storage-item.js";
import * as colorRangeSlider from "./color-range-slider.js";

window.onpageshow = async () => {
    setupAppBar();
    setupPower();

    // store: re-render each time if colors changes "colors"
    window.store.listen("colors", async (data) => {
        await setupColorStorage(data);
    });

    let timeout: NodeJS.Timeout | null = null;
    const onFocus = () => {
        if (timeout !== null) {
            clearTimeout(timeout);
            timeout = null;
        }

        timeout = setTimeout(async () => {
            await window.api.colors();

            timeout = null;
        });
    };
    window.onfocus = () => onFocus();
    onFocus();

    setupRangeSliders();

    console.debug("device address:", page.device());
    console.table({
        activeColor: page.rgbActive(),
        currentColor: page.currentColor(),
        pickedColor: page.color(),
        rangeSliderValues: page.rangeSliderValues(),
    });
};

function setupAppBar(): void {
    const device = page.device();
    const items = window.utils.setupAppBarItems("online-indicator", "title");
    items.title!.innerText = device ? device.server.name : "";
}

async function setupPower() {
    const powerOffBtn =
        document.querySelector<HTMLButtonElement>(`.power button.off`)!;

    powerOffBtn.onclick = async () => {
        window.api.setDevicesColor(
            page.currentColor().map(() => 0),
            page.device(),
        );
    };

    const powerOnBtn =
        document.querySelector<HTMLButtonElement>(`.power button.on`)!;

    powerOnBtn.onclick = async () => {
        window.api.setDevicesColor(page.currentColor(), page.device());
    };
}

async function setupColorStorage(colors: Colors): Promise<void> {
    const colorStorageContainer = document.querySelector<HTMLElement>(
        `.color-storage-container`,
    )!;

    colorStorageContainer.innerHTML = "";

    const currentColor = page.currentColor();
    const currentColorString = currentColor
        .slice(0, 3)
        .join(colorStorageItem.colorSeparator);

    const device = page.device();

    // Create color storage items
    colors.forEach((color, index) => {
        const item = colorStorageItem.create(index, color, {
            device,

            onClick(color): void {
                const colorString = color.join(colorStorageItem.colorSeparator);

                Array.from(colorStorageContainer.children).forEach((child) => {
                    if (child.getAttribute("data-color") === colorString) {
                        if (!child.classList.contains("active")) {
                            child.classList.add("active");

                            window.api.setDevicesColor(
                                [...color, ...page.rangeSliderValues()],
                                device,
                            );
                        }
                    } else {
                        child.classList.remove("active");
                    }
                });
            },

            onChange(color): void {
                if (!item.classList.contains("active")) {
                    return;
                }

                window.api.setDevicesColor(
                    [...color, ...page.rangeSliderValues()],
                    device,
                );
            },

            enableDelete: true,
            async onDelete(_color): Promise<void> {
                await window.api.deleteColor(index);
            },
        });

        if (item.getAttribute("data-color") === currentColorString) {
            item.classList.add("active");
        }

        colorStorageContainer.appendChild(item);
    });

    // New color button
    const newColorBtnInput = document.querySelector<HTMLInputElement>(
        `button.new-color input`,
    )!;

    newColorBtnInput.onchange = () => {
        if (!newColorBtnInput.value) return;
        const color = colorStorageItem.hexToRGB(newColorBtnInput.value);
        window.api.setColor(-1, [...color, ...page.rangeSliderValues()]);
    };
}

async function setupRangeSliders(): Promise<void> {
    const container = document.querySelector<HTMLElement>(".range-sliders")!;
    container.innerHTML = "";

    const device = page.device();

    if ((device.pins || []).length > 3) {
        container.style.display = "block";
    } else {
        container.style.display = "none";
        return;
    }

    if (device.pins) {
        const currentColor = page.currentColor();
        let timeout: NodeJS.Timeout | null = null;

        device.pins.slice(3).forEach((pin, index) => {
            index += 3;
            const slider = colorRangeSlider.create(
                `Pin: ${pin.toString()}`,
                currentColor[index] || 0,
                {
                    async onChange() {
                        if (timeout !== null) {
                            clearTimeout(timeout);
                            timeout = null;
                        }
                        timeout = setTimeout(() => {
                            timeout = null;
                            window.api.setDevicesColor(page.color(), device);
                        }, 250);
                    },
                },
            );

            container.appendChild(slider);
        });
    }
}

const page = {
    address(): string {
        return decodeURIComponent(location.pathname.split("/").reverse()[0]);
    },

    device(): Device {
        const addr = this.address();
        const device = window.store.device(addr);
        if (!device) throw new Error(`device not found for ${addr}`);
        return device;
    },

    currentColor(): Color {
        return (
            window.store.currentDeviceColor(this.address()) ||
            (this.device().pins || []).map(() => 255)
        );
    },

    color(): Color {
        // Get color from active item
        const color = this.rgbActive().slice(0, 3);

        // Get range slider values
        color.push(...this.rangeSliderValues());

        return color;
    },

    rgbActive(): number[] {
        let color = [];
        const activeItem = document.querySelector(`.color-storage-item.active`);
        if (activeItem) {
            color.push(
                ...colorStorageItem.splitDataColor(
                    activeItem.getAttribute("data-color")!,
                ),
            );
        } else {
            color = [255, 255, 255];
        }
        return color;
    },

    rangeSliderValues(): number[] {
        return Array.from(
            document.querySelectorAll<HTMLInputElement>(
                `.range-sliders .color-range-slider input[type="range"]`,
            ),
        ).map((input) => {
            return parseInt(input.value || "0", 10);
        });
    },
};
