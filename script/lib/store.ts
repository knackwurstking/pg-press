export class UIStore extends window.ui.Store<{
    devices: Device[];
    colors: Colors;
    currentDeviceColors: Record<string, Color>;
}> {
    constructor() {
        super("picow-led:");

        this.set("devices", [], true);
        this.set("colors", [], true);

        this.set("currentDeviceColors", {}, true);

        // Update "currentDeviceColors" after any "devices" change
        this.listen(
            "devices",
            async (devices) => {
                devices.forEach((device) => {
                    if (Math.max(...(device.color || [])) > 0) {
                        this.update("currentDeviceColors", (data) => {
                            data[device.server.addr] =
                                device.color ||
                                (device.pins || []).map(() => 255);
                            return data;
                        });
                    }
                });
            },
            true,
        );

        // @ts-expect-error
        this.delete("color"); // TODO: Just to clean up, can be removed before release
    }

    public device(addr: string): Device | null {
        for (const d of this.get("devices") || []) {
            if (d.server.addr === addr) {
                return d;
            }
        }

        return null;
    }

    public currentDeviceColor(addr: string): Color | null {
        return this.get("currentDeviceColors")?.[addr] || null;
    }
}
