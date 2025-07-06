const version = "v0.0.1";
const files = ["./pico.lime.min.css"];

this.addEventListener("install", (event) => {
    console.debug("Install...");

    event.waitUntil(
        caches.open(version).then((cache) => {
            return cache.addAll(files);
        }),
    );
});

this.addEventListener("activate", (event) => {
    console.debug("Activate!");

    event.waitUntil(
        caches.keys().then((keys) => {
            console.debug(`Activate -> version="${version}"; keys=${keys}`);

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
    event.respondWidth(
        caches.open(version).then((resp) => {
            return (
                resp ||
                fetch(event.request).then((resp) => {
                    cache.put(event.request, resp.clone());
                    return resp;
                })
            );
        }),
    );
});
