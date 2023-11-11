// since there's no dynamic data here, we can prerender
// it so that it gets served as a static asset in production
export const prerender = false;

import { invoke } from "@tauri-apps/api/tauri";
import type { ListInfo } from "$lib/types";

export const load = async ({ url }) => {
    let lists: ListInfo[];

    // Retrieve Data
    await invoke("get_lists").then((value) => {
        lists = value;
    });

    return {
        lists,
    };
};
