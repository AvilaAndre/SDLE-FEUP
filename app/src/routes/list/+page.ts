import { invoke } from "@tauri-apps/api/tauri";
import { error } from "@sveltejs/kit";

// just to ignore typescript error
type UrlArg = {
    url: any;
};

export const load = async ({ url }: UrlArg) => {
    const listID: number = parseInt(url.searchParams.get("id"));

    let response;

    let success = false;
    // Retrieve Data
    await invoke("get_shopping_list", { id: listID })
        .then((value: any) => {
            response = value;
            success = true;
        })
        .catch((value: String) => {
            console.log(value);
        });

    if (success) return response;
    else throw error(404, "Failed to retrieve shopping list data");
};
