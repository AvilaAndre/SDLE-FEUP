<script>
    import Tabs from "$lib/components/Tabs.svelte";
    import "../app.css";

    import { tabsList } from "$lib/writables/listTabs";
    import { invoke } from "@tauri-apps/api/tauri";
    import { typewatch } from "../utils/typewatch";

    export let data;

    const updateServerAddress = async () => {
        await invoke("set_server_address", {
            address: data.address.trim(),
        }).then(() => {
            console.log("updated address", data.address);
        });
    };
</script>

<div class="app">
    <section class="bg-gray-200 w-screen h-12 fixed z-20">
        <Tabs bind:tabs={$tabsList} />
    </section>
    <div class="h-12" />
    <main>
        <slot />
    </main>
    <div
        class="fixed right-0 bottom-0 bg-sunglow h-fit w-fit p-1 rounded-tl-md"
    >
        <label for="server-address" class="bg-sunglow p-2">
            Server Address
        </label>
        <input
            id="server-address"
            type="text"
            class="border-2 border-sunglow pl-1"
            bind:value={data.address}
            on:keyup={() =>
                typewatch(() => {
                    updateServerAddress();
                }, 1000)}
        />
    </div>
</div>
