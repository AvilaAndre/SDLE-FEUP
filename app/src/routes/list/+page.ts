import { invoke } from "@tauri-apps/api/tauri";
import { error } from "@sveltejs/kit";
import { crdtToShoppingList } from "$lib/crdt/translator";
import type { CRDTShoppingListData } from "$lib/types";

// just to ignore typescript error
type UrlArg = {
    url: any;
};

export const load = async ({ url }: UrlArg) => {
    const listID: Text = url.searchParams.get("id");

    let crdt: CRDTShoppingListData;
    let success = false;

    // Retrieve Data
    await invoke("get_shopping_list", { id: listID })
        .then((value: any) => {
            crdt = value;
            success = true;
        })
        .catch((value: String) => {
            console.log(value);
        });

    if (success) {
        return crdtToShoppingList(crdt);
    } else throw error(404, "Failed to retrieve shopping list data");
};
