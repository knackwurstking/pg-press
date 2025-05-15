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
                // Main, window
                main: "./script/main.ts",

                // Layout
                "layout/base": "./script/layout/base.ts",

                // Content
                "content/home": "./script/content/home.ts",
            },
            output: {
                dir: "./public/js/",
                entryFileNames: "[name].js",
            },
        },
    },

    define: {
        "process.env.SERVER_PATH_PREFIX": JSON.stringify(
            process.env.SERVER_PATH_PREFIX,
        ),
    },
});
