import { Octokit } from "octokit";

import * as globals from "./globals";
import * as listStores from "./list-stores";
import * as types from "./types";

export async function pull(
    storeKey: keyof listStores.ListStoreData,
    gistID: string,
): Promise<{ gist: types.Gist | null; lists: any[] }> {
    const resp = await getGist(gistID);
    const listsStore = listStores.get(storeKey);
    const newLists: any[] = [];

    for (const file of Object.values(resp || {})) {
        if (!file?.content) continue;

        const data = listsStore.validate(file.content);
        if (data === null) {
            console.error(
                `Invalid data for "${listsStore.title()}":`,
                file.content,
            );
            throw new Error(`ungültige Daten für "${listsStore.title()}"!`);
        }

        newLists.push(data);
    }

    return {
        gist: {
            id: gistID,
            revision: await getRevision(gistID),
            token: globals.store.get(storeKey)!.gist?.token || "",
        },
        lists: newLists,
    };
}

export async function push(
    storeKey: keyof listStores.ListStoreData,
    apiToken: string,
    gistID: string,
): Promise<{ gist: types.Gist | null; lists: any[] }> {
    const storeData = globals.store.get(storeKey)!;
    const ls = listStores.get(storeKey);

    // Take store data and push to gist
    let allowPush = false;
    const data: { [key: string]: { content: string } | null } = {};
    storeData.lists.forEach((list) => {
        data[ls.fileName(list)] = { content: JSON.stringify(list) };
    });

    // Before patching a gist, get the gist and compare + set files no longer
    // existing to null
    const currentGistFiles = await getGist(gistID);
    if (!!currentGistFiles) {
        const newFileNames = Object.keys(data);
        Object.entries(currentGistFiles).forEach(([k, v]) => {
            if (!newFileNames.includes(k)) {
                data[k] = null;
                allowPush = true;
                return;
            }

            // Compare files
            const file = data[k];
            if (file?.content !== v?.content) allowPush = true;
        });

        const currentFileNames = Object.keys(currentGistFiles);
        newFileNames.forEach((name) => {
            if (!currentFileNames.includes(name)) {
                allowPush = true;
            }
        });
    }

    if (allowPush) await patchGist(apiToken, gistID, data);
    else {
        alert("Nichts hat sich geändert!");
    }

    return {
        gist: {
            id: gistID,
            revision: !allowPush
                ? storeData.gist?.revision || null
                : storeData.gist?.revision
                  ? storeData.gist.revision + 1
                  : await getRevision(gistID),
            token: apiToken,
        },
        lists: storeData.lists,
    };
}

export function getRevision(gistID: string, page: number = 1): Promise<number> {
    return new Promise(async (resolve, reject) => {
        const octokit = new Octokit();
        octokit.log.error = (message: string) => {
            return reject(new Error(message));
        };

        // NOTE: Uncomment for development
        //console.debug("fetch revision...");
        //return 0;

        const resp = await octokit.request("GET /gists/{gist_id}/commits", {
            gist_id: gistID,
            headers: {
                "X-GitHub-Api-Version": "2022-11-28",
            },
            per_page: 100,
            page: page,
        });

        if (resp.status !== 200) {
            reject(new Error(`GET ${resp.url} ${resp.status}`));
        }

        if (resp.data.length < 100) {
            return resolve(resp.data.length);
        }

        return resolve(
            (await getRevision(gistID, page + 1)) + resp.data.length,
        );
    });
}

async function getGist(gistID: string): Promise<
    | {
          [key: string]: {
              filename?: string;
              type?: string;
              language?: string;
              raw_url?: string;
              size?: number;
              truncated?: boolean;
              content?: string;
          } | null;
      }
    | undefined
> {
    return new Promise(async (resolve, reject) => {
        const octokit = new Octokit();

        octokit.log.error = (message: string) => {
            return reject(new Error(message));
        };

        try {
            const resp = await octokit.request("GET /gists/{gist_id}", {
                gist_id: gistID,
                headers: {
                    "X-GitHub-Api-Version": "2022-11-28",
                },
            });

            if (resp.status !== 200) {
                return reject(
                    new Error(
                        `anfrage von "GET /gist/${gistID}" ist mit Statuscode ${resp.status} fehlgeschlagen`,
                    ),
                );
            }

            return resolve(resp.data.files);
        } catch (err) {
            return reject(err);
        }
    });
}

async function patchGist(
    apiToken: string,
    gistID: string,
    files: { [key: string]: { content: string } | null },
): Promise<void> {
    return new Promise(async (resolve, reject) => {
        const octokit = new Octokit();

        octokit.log.error = (message: string) => {
            return reject(new Error(message));
        };

        try {
            const resp = await octokit.request("PATCH /gists/{gist_id}", {
                gist_id: gistID,
                description: `Update: ${new Date()}`,
                // @ts-ignore
                files: files,
                headers: {
                    Authorization: `Bearer ${apiToken}`,
                    "X-GitHub-Api-Version": "2022-11-28",
                },
            });

            if (resp.status !== 200) {
                return reject(
                    new Error(
                        `anfrage von "PATCH /gists/${gistID}" ist mit Statuscode ${resp.status} fehlgeschlagen`,
                    ),
                );
            }
        } catch (err) {
            return reject(err);
        }

        return resolve();
    });
}
