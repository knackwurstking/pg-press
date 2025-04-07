import { Directory, Filesystem } from "@capacitor/filesystem";
import { Share } from "@capacitor/share";
import * as jspdf from "jspdf";
import jsPDFAutotable from "jspdf-autotable";
import * as ui from "ui";

import * as dialogs from "../../dialogs";
import * as globals from "../../globals";
import * as listStores from "../../list-stores";
import * as types from "../../types";
import * as utils from "../../utils";
import * as query from "../../utils-query";

interface Param {
    listKey?: string;
}

const tableHeader = [
    "Stärke",
    "Marke (Höhe)",
    "Blech Stempel",
    "Blech Marke",
    "Stf. P2 - P5",
    "Stf. P0",
];

let cleanup: types.CleanUp[] = [];

export async function onMount() {
    const param: Param = ui.router.hash.getSearchParam();

    const metalSheet = globals.getMetalSheet(param.listKey || "");
    if (!metalSheet) {
        throw new Error(`metal sheet "${param.listKey}" not found`);
    }

    // Set the app bar title
    cleanup.push(
        utils.setAppBarTitle(
            metalSheet.data.press > -1
                ? `[P${metalSheet.data.press}] ${metalSheet.format} ${metalSheet.toolID}`
                : `${metalSheet.format} ${metalSheet.toolID}`,
        ),
    );

    metalSheet.data.table.data = sortTableData(metalSheet.data.table.data);

    setupAppBarFilterButton(metalSheet);
    setupAppBarPrinterButton(metalSheet);
    setupAppBarEditSheetButton(metalSheet);
    setupAppBarAddTableEntryButton(metalSheet);
    render(metalSheet);
}

export async function onDestroy() {
    cleanup.forEach((fn) => fn());
    cleanup = [];
}

function setupAppBarFilterButton(metalSheet: types.MetalSheet) {
    const filterButton = query.appBar_ButtonFilter();

    filterButton.onclick = async () => {
        const dialog = dialogs.metalSheetFilter(metalSheet.data.table.filter);

        const data = await dialog.utils?.open();
        if (!data) return;

        metalSheet.data.table.filter = data;

        const ls = listStores.get("metal-sheets");

        try {
            ls.replaceInStore(metalSheet);
        } catch (err) {
            alert(err);
            filterButton.click();
            return;
        }

        reload();
    };

    filterButton.style.display = "inline-flex";

    cleanup.push(() => {
        filterButton.style.display = "none";
        filterButton.onclick = null;
    });
}

function setupAppBarPrinterButton(metalSheet: types.MetalSheet) {
    const printerButton = query.appBar_ButtonPrinter();

    printerButton.onclick = async () => {
        const pdf = new jspdf.jsPDF();
        const listStore = listStores.get("metal-sheets");

        jsPDFAutotable(pdf, {
            head: [
                [
                    {
                        content: `${listStore.listKey(metalSheet)}`,
                        colSpan: globals.metalSheetSlots.length,
                        styles: {
                            fillColor: [255, 255, 255],
                            textColor: [0, 0, 0],
                        },
                    },
                ],
                globals.metalSheetSlots,
            ],
            body: metalSheet.data.table.data,
            theme: "grid",
            styles: {
                valign: "middle",
                halign: "center",
                font: "Courier",
                fontStyle: "bold",
                fontSize: 12,
            },
            headStyles: {
                fillColor: [0, 0, 0],
                textColor: [255, 255, 255],
            },
        });

        const fileName = listStore.fileName(metalSheet).replace(/(\.json)$/, ".pdf");

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

function setupAppBarEditSheetButton(metalSheet: types.MetalSheet) {
    // Enable app bar button for editing the current sheet
    const listEditButton = query.appBar_ButtonListEdit();

    listEditButton.onclick = async () => {
        const dialog = dialogs.metalSheet(metalSheet);

        const data = await dialog.utils?.open();
        if (!data) return;

        if (!dialog.utils!.validate()) {
            listEditButton.click();
            return;
        }

        try {
            const ls = listStores.get("metal-sheets");
            ls.replaceInStore(data, metalSheet);

            if (ls.listKey(data) !== ls.listKey(metalSheet)) {
                ui.router.hash.goTo(
                    {
                        listKey: ls.listKey(data),
                    },
                    "metal-sheets",
                );
            } else {
                reload();
            }
        } catch (err) {
            dialog.query!.inputs.format.ariaInvalid = "";
            dialog.query!.inputs.toolID.ariaInvalid = "";
            alert(err);
            listEditButton.click();
            return;
        }
    };

    listEditButton.style.display = "inline-flex";

    cleanup.push(() => {
        listEditButton.style.display = "none";
        listEditButton.onclick = null;
    });
}

function setupAppBarAddTableEntryButton(metalSheet: types.MetalSheet) {
    const addButton = query.appBar_ButtonAdd();
    addButton.onclick = async () => {
        const data = await dialogs.metalSheetTableEntry().utils!.open();

        if (!data) {
            return;
        }

        metalSheet.data.table.data.push(data);

        const ls = listStores.get("metal-sheets");
        ls.replaceInStore(metalSheet);
        reload();
    };
    addButton.style.display = "inline-flex";
    cleanup.push(() => {
        addButton.style.display = "none";
        addButton.onclick = null;
    });
}

function render(metalSheet: types.MetalSheet) {
    const el = routerTargetElements();

    el.thead.innerHTML = "";
    el.tbody.innerHTML = "";

    const tr = document.createElement("tr");
    el.thead.appendChild(tr);
    tableHeader.forEach((head, index) => {
        if (!!metalSheet.data.table.filter?.includes(index)) {
            return;
        }

        const el = document.createElement("th");
        el.style.textWrap = "pretty";
        el.innerText = head;
        tr.appendChild(el);
    });

    metalSheet.data.table.data.forEach((row) => {
        const tr = document.createElement("tr");
        el.tbody.appendChild(tr);

        tr.role = "button";
        tr.classList.add("ui-none-select");
        tr.style.cursor = "pointer";

        // Delete row on right click
        tr.oncontextmenu = (e) => {
            e.preventDefault();

            if (confirm(`You want to delete this item: "${row.join(",")}" ?`)) {
                const rowJoin = row.join(",");
                metalSheet.data.table.data = metalSheet.data.table.data.filter(
                    (r) => r.join(",") !== rowJoin,
                );

                const ls = listStores.get("metal-sheets");
                ls.replaceInStore(metalSheet);
                reload();
            }
        };

        // Edit row on click
        tr.onclick = async () => {
            const data = await dialogs.metalSheetTableEntry(row).utils!.open();
            if (!data) {
                return;
            }

            // Update data and re-render table
            const rowJoin = row.join(",");
            metalSheet.data.table.data = metalSheet.data.table.data.map((r) => {
                if (r.join(",") === rowJoin) {
                    return data;
                }

                return r;
            });

            const ls = listStores.get("metal-sheets");
            ls.replaceInStore(metalSheet);
            reload();
        };

        row.forEach((cell, index) => {
            if (!!metalSheet.data.table.filter?.includes(index)) {
                return;
            }

            const el = document.createElement("td");
            el.innerText = cell;
            tr.appendChild(el);
        });
    });
}

async function reload() {
    await onDestroy();
    await onMount();
}

function sortTableData(data: string[][]): string[][] {
    data.sort((a, b) => {
        const thicknessA = parseFloat(a[0]);
        const thicknessB = parseFloat(b[0]);

        if (thicknessA > thicknessB) {
            // Compare thickness
            return 1;
        } else if (thicknessA === thicknessB && parseFloat(a[1]) > parseFloat(b[1])) {
            // Compare height if thickness is event
            return 1;
        }

        return -1;
    });

    return data;
}

function routerTargetElements() {
    const rt = query.routerTarget();

    return {
        thead: rt.querySelector(`table thead`)!,
        tbody: rt.querySelector(`table tbody`)!,
    };
}
