// since there's no dynamic data here, we can prerender
// it so that it gets served as a static asset in production
export const prerender = true;

import { invoke } from "@tauri-apps/api/tauri";

export const load = async ({ url }) => {
    type ListInfo = {
        title: string;
        id: string;
        shared: boolean;
    };

    let lists: ListInfo[];

    // Retrieve Data
    await invoke("get_lists").then((value) => {
        lists = value;
    });

    return {
        lists,
    };
};
