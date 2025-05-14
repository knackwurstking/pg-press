const CACHE_VERSION = process.env.VERSION;
const CURRENT_CACHE = "picow-led-" + CACHE_VERSION;

const cacheFiles = [
    process.env.SERVER_PATH_PREFIX + "/",
    process.env.SERVER_PATH_PREFIX + "/settings",

    process.env.SERVER_PATH_PREFIX + "/css/style.css",
    process.env.SERVER_PATH_PREFIX + "/css/ui-v4.2.0.css",

    process.env.SERVER_PATH_PREFIX + "/js/devices.js",
    process.env.SERVER_PATH_PREFIX + "/js/devices-address.js",
    process.env.SERVER_PATH_PREFIX + "/js/layout-base.js",
    process.env.SERVER_PATH_PREFIX + "/js/main.js",
    process.env.SERVER_PATH_PREFIX + "/js/settings.js",
    process.env.SERVER_PATH_PREFIX + "/js/ui-v4.2.0.min.umd.cjs",

    process.env.SERVER_PATH_PREFIX + "/icons/apple-touch-icon-180x180.png",
    process.env.SERVER_PATH_PREFIX + "/icons/favicon.ico",
    process.env.SERVER_PATH_PREFIX + "/icons/icon.png",
    process.env.SERVER_PATH_PREFIX + "/icons/maskable-icon-512x512.png",
    process.env.SERVER_PATH_PREFIX + "/icons/pwa-192x192.png",
    process.env.SERVER_PATH_PREFIX + "/icons/pwa-512x512.png",
    process.env.SERVER_PATH_PREFIX + "/icons/pwa-64x64.png",

    "https://fonts.googleapis.com/css2?family=Recursive:slnt,wght,CASL,CRSV,MONO@-15..0,300..1000,0..1,0..1,0..1&display=swap",
];

// NOTE: This will be prefixed with ".*" and suffixed with "$"
const blackList = [
    "/api/.*",
    "/screenshots/.*",
    "/service-worker.js",
    "/manifest.json",
];

self.addEventListener("activate", (evt) => {
    // @ts-expect-error
    evt.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cacheName) => {
                    if (cacheName !== CURRENT_CACHE) {
                        return caches.delete(cacheName);
                    }
                }),
            );
        }),
    );
});

self.addEventListener("install", (evt) => {
    // @ts-expect-error
    evt.waitUntil(
        (() => {
            caches
                .keys()
                .then((keyList) => {
                    return Promise.all(
                        keyList.map(function (key) {
                            return caches.delete(key);
                        }),
                    );
                })
                .then(() => {
                    caches.open(CURRENT_CACHE).then((cache) => {
                        return cache.addAll(cacheFiles);
                    });
                });
        })(),
    );
});

const fromNetwork = (
    /** @type {Request} */ request,
    /** @type {number} */ timeout,
) => {
    return new Promise((resolve, reject) => {
        const timeoutId = setTimeout(reject, timeout);

        fetch(request).then((response) => {
            clearTimeout(timeoutId);
            update(request, response.clone());
            resolve(response);
        }, reject);
    });
};

const fromCache = async (/** @type {Request} */ request) => {
    const cache = await caches.open(CURRENT_CACHE);
    const matching = await cache.match(request);
    return matching || cache.match("/offline/");
};

const update = async (
    /** @type {Request} */ request,
    /** @type {Response} */ response,
) => {
    const cache = await caches.open(CURRENT_CACHE);
    cache.put(request, response);
};

const isBlackListed = (/** @type {string} */ url) => {
    return blackList.find((path) => new RegExp(".*" + path + "$").test(url));
};

self.addEventListener(
    "fetch",
    (/** @type {Event & { request: Request }} */ evt) => {
        if (!isBlackListed(evt.request.url)) {
            console.debug(`fetch: caching: ${evt.request.url}`);

            // @ts-expect-error
            evt.respondWith(
                fromNetwork(evt.request, 1e4).catch(() =>
                    fromCache(evt.request),
                ),
            );
        } else {
            console.debug(`fetch: no caching for: ${evt.request.url}`);
        }
    },
);
