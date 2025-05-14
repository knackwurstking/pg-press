import { defineConfig } from "vite";
import { babel } from "@rollup/plugin-babel";
import { viteRequire } from "vite-require";

export default defineConfig({
    plugins: [
        babel({
            babelHelpers: "bundled",
            presets: [
                [
                    "@babel/preset-env",
                    {
                        targets: {
                            browsers: ["chrome >= 55"],
                        },
                    },
                ],
            ],
        }),
        viteRequire(),
    ],

    build: {
        minify: false,
        copyPublicDir: false,
        emptyOutDir: false,
        rollupOptions: {
            input: {
                "service-worker": "./script/pwa/service-worker.js",
            },
            output: {
                dir: "./public/",
                entryFileNames: "[name].js",
            },
        },
    },

    //esbuild: {
    //    legalComments: "inline",
    //},

    define: {
        "process.env.SERVER_PATH_PREFIX": JSON.stringify(
            process.env.SERVER_PATH_PREFIX,
        ),
        "process.env.VERSION": JSON.stringify(process.env.VERSION),
    },
});
