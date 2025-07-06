const version = "v1";
const files = ["./pico.lime.min.css"];

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
                        // @ts-expect-error - request not exists
                        cache.put(event.request, resp.clone());
                        return resp;
                    })
                );
            });
        }),
    );
});
