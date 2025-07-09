// Version History:
//
//  - v0: Initial version
//  - v0.1: Updated files
//  - v0.2: Updated files, Add bootstrap-icons (woff, woff2, css)
//  - v0.3: Updated files, ./css/style.css removed
//  - v0.4: Updated files, Add ui-dev.min.css, manifest.json

const version = "v0.4";

const files = [
    "./apple-touch-icon-180x180.png",
    "./bootstrap-icons.woff",
    "./bootstrap-icons.woff2",
    "./favicon.ico",
    "./icon.png",
    "./manifest.json",
    "./maskable-icon-512x512.png",
    "./pwa-192x192.png",
    "./pwa-512x512.png",
    "./pwa-64x64.png",
    "./ui-dev.min.css",

    "./css/bootstrap-icons.min.css",
];

this.addEventListener("install", (event) => {
    console.debug("Install...", { files });

    // @ts-expect-error - waitUntil not exists
    event.waitUntil(
        caches.open(version).then((cache) => {
            return cache.addAll(files);
        }),
    );
});

this.addEventListener("activate", (event) => {
    console.debug("Activate!");

    // @ts-expect-error - waitUntil not exists
    event.waitUntil(
        caches.keys().then((keys) => {
            console.debug(`Activate ->`, { version, keys });

            // @ts-expect-error - wrong target library (tsconfig)
            return Promise.all(
                keys
                    .filter((key) => {
                        return key !== version;
                    })
                    .map((key) => {
                        return caches.delete(key);
                    }),
            );
        }),
    );
});

// NOTE: Cache first
this.addEventListener("fetch", (event) => {
    // @ts-expect-error - respondWith not exists
    event.respondWith(
        caches.open(version).then((cache) => {
            // @ts-expect-error - request not exists
            return cache.match(event.request).then((resp) => {
                if (resp) {
                    // @ts-expect-error - request not exists
                    console.debug("Fetch:", event.request, { resp });
                }

                return (
                    resp ||
                    // @ts-expect-error - request not exists
                    fetch(event.request).then((resp) => {
                        //cache.put(event.request, resp.clone());
                        return resp;
                    })
                );
            });
        }),
    );
});
