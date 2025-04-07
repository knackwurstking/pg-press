import * as types from "./types";

export function toVis(dataString: string): types.Vis | null {
    /** @type {PGStore_Vis} */
    const vis: types.Vis = (() => {
        const date = new Date();
        const m = (date.getMonth() + 1).toString().padStart(2, "0");
        const d = date.getDate().toString().padStart(2, "0");

        return {
            date: date.getTime(),
            title: `${date.getFullYear()}-${m}-${d}`,
            data: [],
        };
    })();

    for (let line of dataString.split("\n")) {
        line = line.trim();
        if (!line) continue;

        const product: types.Product = {
            lotto: "",
            name: "",
            format: "",
            stamp: "",
            thickness: -1,
        };

        const lineTabSplit = line.split("\t");

        product.lotto = lineTabSplit[0] || "";
        if (product.lotto === "") {
            throw new Error(`product lotto missing in ${line}`);
        }

        product.name = lineTabSplit[1] || "";
        if (product.name === "") {
            throw new Error(`product name not found in "${line}"`);
        }

        product.format = lineTabSplit[2] || "";
        if (product.format === "") {
            throw new Error(`product format not found in "${line}"`);
        }
        product.format = fixFormatString(product.format);

        product.thickness = parseFloat(lineTabSplit[3].replace(/,/g, ".") || "-1");
        if (product.thickness === -1) {
            throw new Error(`product thickness not found in "${line}"`);
        }

        product.stamp = lineTabSplit[4] || "";
        if (product.stamp === "") {
            throw new Error(`product stamp not found in "${line}"`);
        }

        vis.data.push(product);
    }

    return vis;
}

export function fixFormatString(format: string): string {
    try {
        const formatSplit = format.split(/[xX]/i);
        if (formatSplit.length > 2) return format;

        const wSuffix = formatSplit[0].match(/[0-9]+(.*)/i)?.[1] || "";
        const hSuffix = formatSplit[1].match(/[0-9]+(.*)/i)?.[1] || "";

        const w = parseFloat(formatSplit[0].replace(/,/g, "."));
        const h = parseFloat(formatSplit[1].replace(/,/g, "."));
        format = w > h ? `${w}${wSuffix}X${h}${hSuffix}` : `${h}${wSuffix}X${w}${hSuffix}`;
    } catch (err) {
        console.warn(`Fix product format "${format}": ${err}`);
    }

    return format;
}

export function textToHTML(content: string) {
    return content.trim().replace(/\n/g, "<br/>");
}
