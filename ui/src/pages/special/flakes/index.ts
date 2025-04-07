import { Directory, Filesystem } from "@capacitor/filesystem";
import { Share } from "@capacitor/share";
import * as jspdf from "jspdf";
import jsPDFAutotable from "jspdf-autotable";
import * as ui from "ui";

import * as dialogs from "../../../dialogs";
import * as globals from "../../../globals";
import * as listStores from "../../../list-stores";
import * as types from "../../../types";
import * as utils from "../../../utils";
import * as query from "../../../utils-query";
import * as create from "./create";

interface Param {
    listKey?: string;
}

let cleanup: types.CleanUp[] = [];

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const flakes = globals.store.get("special")!.lists.find((list) => list.type === "flakes");
    if (!flakes) {
        throw new Error(`flakes special list "${param.listKey}" not found`);
    }

    cleanup.push(utils.setAppBarTitle(flakes.title));

    setupAppBarPrinterButton(flakes);
    setupAppBarAddButton(flakes);
    render(flakes);
}

export async function onDestroy() {
    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function setupAppBarPrinterButton(flakes: types.SpecialFlakes) {
    const printerButton = query.appBar_ButtonPrinter();

    printerButton.onclick = async () => {
        const pdf = new jspdf.jsPDF();
        const ls = listStores.get("special");

        const createTable = (
            press: types.SpecialFlakes_PressSlot,
            entries: types.SpecialFlakesEntry[],
        ) => {
            const bodyData = [];
            for (const entry of entries) {
                const slots = [
                    entry.compatatore,
                    `${entry.primary.percent}%\n${entry.primary.value}`,
                    ``,
                    ``,
                    ``,
                    ``,
                    ``,
                    ``,
                ];

                for (const consumption of entry.secondary) {
                    slots[globals.flakesTowerSlots.indexOf(consumption.slot) + 2] =
                        `${consumption.percent}%\n${consumption.value}`;
                }

                bodyData.push(slots);
            }

            jsPDFAutotable(pdf, {
                head: [
                    [
                        {
                            content: `${ls.listKey(flakes)} - ${globals.flakesPressSlotsFull[globals.flakesPressSlots.indexOf(press)]}`,
                            colSpan: globals.flakesTowerSlots.length + 2, // + "C1" + "Main"
                            styles: {
                                fillColor: [255, 255, 255],
                                textColor: [0, 0, 0],
                            },
                        },
                    ],
                    ["C1", "Main", ...globals.flakesTowerSlots],
                ],
                body: bodyData,
                theme: "grid",
                styles: {
                    valign: "middle",
                    halign: "center",
                    font: "Courier",
                    fontSize: 12,
                },
                headStyles: {
                    fillColor: [0, 0, 0],
                    textColor: [255, 255, 255],
                },
            });
        };

        globals.flakesPressSlots.forEach((slot, index) => {
            createTable(
                slot,
                flakes.data.filter((e) => e.press === slot),
            );

            if (index < globals.flakesPressSlots.length - 1) {
                pdf.addPage();
            }
        });

        const fileName = ls.fileName(flakes).replace(/(\.json)$/, ".pdf");

        if (process.env.CAPACITOR) {
            Share.share({
                title: fileName,
                dialogTitle: fileName,
                url: (
                    await Filesystem.writeFile({
                        path: fileName,
                        // @ts-ignore
                        data: pdf.output("datauri"),
                        directory: Directory.Cache,
                    })
                ).uri,
            });
        } else {
            pdf.save(fileName);
        }
    };

    printerButton.style.display = "inline-flex";

    cleanup.push(() => {
        printerButton.style.display = "none";
        printerButton.onclick = null;
    });
}

function setupAppBarAddButton(flakes: types.SpecialFlakes) {
    const addButton = query.appBar_ButtonAdd();

    addButton.onclick = async () => {
        const data = await dialogs.specialFlakesEntry().utils!.open();

        if (!data) return;
        flakes.data.push(data);

        const ls = listStores.get("special");
        ls.replaceInStore(flakes);
        reload();
    };

    addButton.style.display = "inline-flex";

    cleanup.push(() => {
        addButton.style.display = "none";
        addButton.onclick = null;
    });
}

function render(flakes: types.SpecialFlakes) {
    const el = routerTargetElements();

    el.tableContainer.innerHTML = "";

    globals.flakesPressSlots
        .map((slot, index) =>
            create.table(
                globals.flakesPressSlotsFull[index],
                flakes.data.filter((entry) => entry.press === slot),
                async (entries) => {
                    flakes.data = [
                        ...flakes.data.filter((entry) => entry.press !== slot),
                        ...entries,
                    ];

                    const ls = listStores.get("special");
                    ls.replaceInStore(flakes);
                    reload();
                },
            ),
        )
        .forEach((table) => {
            el.tableContainer.appendChild(table.element);
            cleanup.push(table.destroy);
        });
}

async function reload() {
    await onDestroy();
    await onMount();
}

function routerTargetElements(): {
    tableContainer: HTMLUListElement;
} {
    const rt = query.routerTarget();

    return {
        tableContainer: rt.querySelector(`.table-container`)!,
    };
}
