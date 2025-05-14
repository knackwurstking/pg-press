function setupAppBar(): void {
    const items = window.utils.setupAppBarItems("online-indicator", "title");
    items["title"]!.innerText = "Settings";
}

window.addEventListener("pageshow", () => {
    setupAppBar();
});
