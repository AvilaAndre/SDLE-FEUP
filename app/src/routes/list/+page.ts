import { invoke } from "@tauri-apps/api/tauri";

export const load = async ({ url }) => {
    let listTitle: string;
    const listID: string = url.searchParams.get("id");
    let listItems: string[];

    type ListData = {
        title: string;
        items: string[];
    };

    // Retrieve Data
    await invoke("get_mock_data", { id: listID }).then((value: ListData) => {
        listTitle = value.title;
        listItems = value.items;
    });

    return {
        title: listTitle || "Error Loading",
        id: listID,
        items: listItems || [],
    };
};
