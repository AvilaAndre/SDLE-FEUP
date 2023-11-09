<script lang="ts">
    import ShareIcon from "$lib/icons/ShareIcon.svelte";
    import SyncIcon from "$lib/icons/SyncIcon.svelte";
    import UploadIcon from "$lib/icons/UploadIcon.svelte";
    import DownloadIcon from "$lib/icons/DownloadIcon.svelte";

    type ListPageData = {
        title: string;
        id: string;
        items: string[];
    };

    export let data: ListPageData;

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
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<div class="flex flex-col justify-start items-center w-full">
    <div class="mt-3 mb-6 w-full grid grid-flow-row grid-cols-[1fr_1fr]">
        <div>
            {#if !published}
                <p>Nothing here yet</p>
            {:else}
                <button
                    type="button"
                    on:click={shareShoppingList}
                    class="flex flex-row bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm gap-x-1"
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
                    class="flex flex-row bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm gap-x-1"
                >
                    <ShareIcon className="w-6" />
                    <p>Share</p>
                </button>
            {/if}
            {#if published}
                <div class="inline-flex gap-1">
                    <p>
                        Last updated {lastUpdate} minutes ago
                    </p>
                    <button
                        type="button"
                        on:click={syncShoppingList}
                        class="flex flex-row bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm"
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
                        class="flex flex-row bg-transparent transition-colors p-1 rounded-sm hover:bg-gray-300 disabled:opacity-50 disabled:bg-gray-300"
                        disabled
                    >
                        <UploadIcon className="w-6" />
                    </button>
                </div>
            {/if}
        </div>
    </div>
    <div class="h-full w-[36rem]">
        <div>
            <h1 class="text-4xl">{data.title}</h1>
            <h4 class="text-sm text-slate-700">{data.id}</h4>
        </div>
        <br />
        <ul>
            {#each data.items as item}
                <li class="text-lg pl-2 hover:bg-gray-100 hover:cursor-pointer">{item}</li>
            {/each}
        </ul>
    </div>
</div>
