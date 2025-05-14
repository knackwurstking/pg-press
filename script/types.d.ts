export declare global {
    interface Window {
        ui: typeof import("ui");
        store: import("./lib/store").UIStore;
        api: import("./lib/api").Api;
        utils: Utils;
        ws: import("ui").WS<WSMessageData>;
    }

    declare type Utils = {
        setupAppBarItems: (...itemNames: AppBarItemName[]) => AppBarItems;
        setOnlineIndicatorState: (state: boolean) => void;
        registerServiceWorker: () => void;
    };

    declare type AppBarItemName =
        | "online-indicator"
        | "title"
        | "settings-button";

    declare type AppBarItems = Partial<Record<AppBarItemName, HTMLElement>>;

    declare type Device = {
        server: Server;
        online: boolean;
        error: string;
        color: Color | null;
        pins: Pins | null;
    };

    declare type Server = {
        addr: string;
        name: string;
    };

    declare type Color = number[];
    declare type Pins = number[];

    declare type Colors = Color[];

    declare type WSMessageData =
        | {
              type: "device";
              data: Device;
          }
        | {
              type: "colors";
              data: Colors;
          };
}
