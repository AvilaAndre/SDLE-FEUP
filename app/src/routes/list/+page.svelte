<script lang="ts">
    import ShareIcon from "$lib/icons/ShareIcon.svelte";
    import SyncIcon from "$lib/icons/SyncIcon.svelte";
    import UploadIcon from "$lib/icons/UploadIcon.svelte";
    import DownloadIcon from "$lib/icons/DownloadIcon.svelte";
    import PublishIcon from "$lib/icons/PublishIcon.svelte";
    import type { ShoppingListData } from "$lib/types";
    import { invoke } from "@tauri-apps/api/tauri";

    export let data: ShoppingListData;

    let nextItem: any;
    let nextItemValue: string;

    let lastUpdate: number = 20;

    let published: boolean = false;
    let hasDataToUpdate: boolean = true;

    const syncShoppingList = () => {
        // TODO: Sync Shopping List logic
        console.log("sync");

        if (hasDataToUpdate) hasDataToUpdate = false;
    };

    const publishShoppingList = () => {
        // TODO: Publish Shopping list logic
        published = true;
        console.log("published");
    };

    const uploadShoppingList = () => {
        // TODO: Upload Shopping List logic
        console.log("upload");
    };

    const shareShoppingList = () => {
        // TODO: Share Shopping List logic
        console.log("share");
    };

    const selectNextItem = () => {
        // This line prevents the nextItem text from being selected when pressing spacebar while writing
        if (document.activeElement?.id != nextItem.id) nextItem.select();
    };

    const addNewItem = async () => {
        await invoke("add_item_to_list", {
            listId: data.list_info.list_id,
            name: nextItemValue,
            qtd: 1,
        }).then((value: any) => {
            if (value) {
                data.items.push({
                    id: 0,
                    name: nextItemValue,
                    list_id: data.list_info.list_id,
                    qtd: 1,
                });
                // to activate svelte's reactivity
                data.items = data.items;
            }
        });

        nextItemValue = "";
    };
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<div class="flex flex-col justify-start items-center w-full">
    <div
        class="px-3 mt-3 mb-6 w-full h-8 grid grid-flow-row grid-cols-[1fr_1fr] items-center"
    >
        <div>
            {#if !published}
                <p>Nothing here yet</p>
            {:else}
                <button
                    type="button"
                    on:click={shareShoppingList}
                    class="flex flex-row items-center bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm gap-x-1"
                >
                    <ShareIcon className="w-6" />
                    <p>Share</p>
                </button>
            {/if}
        </div>
        <div class="flex flex-row justify-end">
            {#if !published}
                <button
                    type="button"
                    on:click={publishShoppingList}
                    class="flex flex-row items-center bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm gap-x-1"
                >
                    <PublishIcon className="w-6" />
                    <p>Publish</p>
                </button>
            {/if}
            {#if published}
                <div class="inline-flex gap-1 items-center">
                    <p>
                        Last updated {lastUpdate} minutes ago
                    </p>
                    <button
                        type="button"
                        on:click={syncShoppingList}
                        class="flex flex-row items-center bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm"
                    >
                        {#if hasDataToUpdate}
                            <DownloadIcon className="w-6" />
                        {:else}
                            <SyncIcon className="w-6 animate-spin" />
                        {/if}
                    </button>

                    <button
                        type="button"
                        on:click={uploadShoppingList}
                        class="flex flex-row items-center bg-transparent transition-colors p-1 rounded-sm hover:bg-gray-300 disabled:opacity-50 disabled:bg-gray-300"
                        disabled
                    >
                        <UploadIcon className="w-6" />
                    </button>
                </div>
            {/if}
        </div>
    </div>
    <div class="h-fit w-full">
        <div class="w-[36rem] mx-auto">
            <h1 class="text-4xl">{data.list_info.title}</h1>
            {#if data.list_info.share_id}
                <h4 class="text-sm text-slate-700 pl-1">
                    {data.list_info.share_id}
                </h4>
            {/if}
        </div>
        <br />
        <ul>
            {#each data.items as item}
                <li class="w-full group">
                    <div
                        class="text-lg w-[36rem] mx-auto pl-2 group-hover:bg-gray-100 hover:cursor-pointer"
                    >
                        {item.name}
                    </div>
                </li>
            {/each}
            <form on:submit|preventDefault={addNewItem}>
                <button
                    type="button"
                    class="w-full cursor-text"
                    on:click={selectNextItem}
                >
                    <input
                        type="text"
                        bind:this={nextItem}
                        bind:value={nextItemValue}
                        name="newItem"
                        id="newItem"
                        class="text-lg w-[36rem] pl-2 hidden-placeholder focus-visible:outline-none"
                        placeholder="Input new item name"
                    />
                </button>
            </form>
        </ul>
    </div>
    <button
        class="h-full w-full min-h-[30vh] cursor-text"
        on:click={selectNextItem}
    />
</div>
