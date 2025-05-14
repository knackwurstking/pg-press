export class Api {
    public async devices(): Promise<Device[]> {
        const url = this.getURL("/api/devices?cache=true");

        let data: any;

        try {
            const resp = await fetch(url);

            try {
                data = await this.handleResponse(resp, url);
            } catch (err) {
                this.fetchResponseError(url, err);
            }
        } catch (err) {
            this.fetchError(url, err);
        }

        if (data) {
            window.store.set("devices", data);
        }

        return data || window.store.get("devices") || [];
    }

    public async setDevicesColor(
        color?: Color | null,
        ...devices: Device[]
    ): Promise<void> {
        const url = this.getURL("/api/devices/color?force=true");

        try {
            await fetch(url, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    devices,
                    color: color || [255, 255, 255, 255],
                }),
            });
        } catch (err) {
            this.fetchError(url, err);
        }
    }

    public async colors(): Promise<Colors> {
        const url = this.getURL("/api/colors");

        let data: any;

        try {
            const resp = await fetch(url);

            try {
                data = await this.handleResponse(resp, url);
            } catch (err) {
                this.fetchResponseError(url, err);
            }
        } catch (err) {
            this.fetchError(url, err);
        }

        if (data) {
            window.store.set("colors", data);
        }

        return data || window.store.get("colors") || [];
    }

    public async color(index: number): Promise<Color | null> {
        const url = this.getURL(`/api/colors/${index}`);

        let data: Color | undefined = undefined;

        try {
            const resp = await fetch(url);

            try {
                data = await this.handleResponse(resp, url);
            } catch (err) {
                this.fetchResponseError(url, err);
            }
        } catch (err) {
            this.fetchError(url, err);
        }

        if (data) {
            window.store.update("colors", (colors) => {
                return colors.map((c, i) => (i === index ? data : c));
            });
        }

        return data || (window.store.get("colors") || [])[index] || null;
    }

    public async setColor(index: number, color: Color): Promise<void> {
        const url = this.getURL(`/api/colors/${index}`);

        try {
            await fetch(url, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(color),
            });
        } catch (err) {
            this.fetchError(url, err);
        }
    }

    public async deleteColor(index: number): Promise<void> {
        const url = this.getURL(`/api/colors/${index}`);

        try {
            await fetch(url, { method: "DELETE" });
        } catch (err) {
            this.fetchError(url, err);
        }
    }

    private fetchError(url: string, err: any) {
        console.error(`fetch ${url}:`, err);
    }

    private fetchResponseError(url: string, err: any) {
        console.error(`Handle fetch response for ${url}:`, err);
    }

    private getURL(path: string): string {
        return process.env.SERVER_PATH_PREFIX + `${path}`;
    }

    private async handleResponse(resp: Response, url: string): Promise<any> {
        const status = resp.status;

        if (!resp.ok) {
            throw new Error(`${status}: ${(await resp.text()) || "???"}`);
        }

        const respData = await resp.json();
        console.debug(`Got data from "${url}":`, respData);
        return respData;
    }
}
