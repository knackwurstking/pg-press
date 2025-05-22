export declare global {
    interface Window {
        ui: typeof import("ui");
    }

    //interface Component<T = any> {
    //    render(target: HTMLElement): Promise<void>;
    //    update(data: T): Promise<void>;
    //    destroy(): Promise<void>;
    //}
}
