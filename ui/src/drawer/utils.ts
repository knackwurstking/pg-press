import * as query from "../utils-query";

const drawer = query.drawer();

export function open() {
    drawer.setAttribute("open", "");
}

export function close() {
    drawer.removeAttribute("open");
}
