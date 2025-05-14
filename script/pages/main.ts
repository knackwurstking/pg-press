import { UIStore } from "../lib/store";
import { Api } from "../lib/api";
import * as utils from "../lib/window/utils";

window.store = new UIStore();
window.api = new Api();
window.utils = utils.create();
window.ws = new window.ui.WS<WSMessageData>(
    (process.env.SERVER_PATH_PREFIX || "") + "/ws",
);
