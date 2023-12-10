import { invoke } from "@tauri-apps/api/tauri";

export const prerender = true;
export const ssr = false;

export const load = async () => {
    let address = "localhost:9988";

    // Retrieve Data
    await invoke("get_server_address")
        .then((value: any) => {
            address = value;
        })
        .catch(() => {
            console.log("failed to get server address");
        });

    return { address };
};
