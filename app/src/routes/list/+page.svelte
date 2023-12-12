<script lang="ts">
    import ShareIcon from "$lib/icons/ShareIcon.svelte";
    import SyncIcon from "$lib/icons/SyncIcon.svelte";
    import UploadIcon from "$lib/icons/UploadIcon.svelte";
    import DownloadIcon from "$lib/icons/DownloadIcon.svelte";
    import PublishIcon from "$lib/icons/PublishIcon.svelte";
    import ProgressWheel from "$lib/icons/ProgressWheel.svelte";
    import type { ListItemInfo, ShoppingListData } from "$lib/types";
    import { invoke } from "@tauri-apps/api/tauri";
    import { closeTab, openTab } from "$lib/writables/listTabs";
    import ListItem from "$lib/components/ListItem.svelte";
    import { typewatch } from "../../utils/typewatch";
    import { crdtToShoppingList } from "$lib/crdt/translator";

    export let data: ShoppingListData;

    let shareIdElement: any;

    let nextItem: any;
    let nextItemValue: string;

    let synchronizingList: boolean = false;

    let uploadingList: boolean = false;

    let publishing: boolean = false;

    const syncShoppingList = () => {
        if (synchronizingList) return;
        synchronizingList = true;

        invoke("sync_list", { listId: data.list_info.list_id })
            .then((value) => {
                if (true) {
                    invoke("get_shopping_list", { id: data.list_info.list_id })
                        .then((value: any) => {
                            data = crdtToShoppingList(value);
                        })
                        .catch((value: String) => {
                            console.log(
                                "failed to retrive list in order to update page:",
                                value
                            );
                        });
                }
            })
            .catch((reason) => console.log("failed to sync:", reason))
            .finally(() => (synchronizingList = false));
    };

    const publishShoppingList = async () => {
        if (publishing) return;

        publishing = true;

        invoke("publish_list", {
            listId: data.list_info.list_id,
        })
            .then((value: any) => {
                if (value) data.list_info.shared = true;
                publishing = false;
            })
            .catch((err) => {
                console.log("error publishing", err);
                publishing = false;
            });
    };

    const uploadShoppingList = () => {
        if (uploadingList) return;
        uploadingList = true;

        //TODO: There is no difference between upload and publish yet
        invoke("publish_list", { listId: data.list_info.list_id })
            .then((value) => {
                console.log("upload success:", value);
            })
            .catch((reason) => console.log("failed to upload:", reason))
            .finally(() => (uploadingList = false));
    };

    const deleteShoppingList = async () => {
        invoke("delete_list", { listId: data.list_info.list_id })
            .then((value) => {
                if (value) {
                    closeTab("/list?id=" + data.list_info.list_id);
                } else {
                    console.log("failed to delete list");
                }
            })
            .catch((reason) => console.log("failed to delete list:", reason));
    };

    const selectNextItem = () => {
        // This line prevents the nextItem text from being selected when pressing spacebar while writing
        if (document.activeElement?.id != nextItem.id) nextItem.select();
    };

    const addNewItem = async () => {
        nextItemValue = nextItemValue.trim();
        if (nextItemValue === "") return;

        // check if new item already exists
        for (let index = 0; index < data.items.length; index++) {
            const item: ListItemInfo = data.items[index];

            if (item.name == nextItemValue) {
                console.log("repeated");
                // TODO: User warnings
                return;
            }
        }

        await invoke("add_item_to_list", {
            listId: data.list_info.list_id,
            name: nextItemValue,
            qtd: 0,
        })
            .then((value: any) => {
                if (value) {
                    data.items.push({
                        name: nextItemValue,
                        checked: false,
                        qtd: 0, //TODO: check this ?
                    });
                    // to activate svelte's reactivity
                    data.items = data.items;
                }
            })
            .catch((err) => {
                console.log("error", err);
            });

        nextItemValue = "";
    };

    const updateListTitle = async () => {
        if (data.list_info.title == "") return;
        await invoke("update_list_title", {
            listId: data.list_info.list_id,
            title: data.list_info.title,
        }).then((value: any) => {
            openTab(data.list_info.title, "/list?id=" + data.list_info.list_id);
        });
    };

    const updateItemCounter = async (
        item: ListItemInfo
    ): Promise<ListItemInfo> => {
        await invoke("update_list_item", {
            listId: data.list_info.list_id,
            listItem: item.name,
            counter: item.qtd,
            checked: item.checked,
        })
            .then((value) => {
                item.checked = value.checked;
                item.qtd = value.counter;
            })
            .catch((error) => {
                console.log(error);
            });

        return item;
    };

    openTab(data.list_info.title, "/list?id=" + data.list_info.list_id);
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<div class="flex flex-col justify-start items-center w-full">
    <div
        class="bg-white px-3 mb-6 w-full h-8 grid grid-flow-row grid-cols-[1fr_0.5fr_1fr] items-center py-2 fixed z-10"
    >
        <div>
            <button
                type="button"
                on:click={deleteShoppingList}
                class="flex flex-row items-center bg-transparent hover:bg-red-300 transition-colors p-1 px-2 rounded-sm gap-x-1"
            >
                <PublishIcon className="w-6" />
                <p>Delete</p>
            </button>
        </div>
        <div class="text-center">
            <h3>
                {data.list_info.title}
            </h3>
        </div>
        <div class="flex flex-row justify-end">
            {#if !data.list_info.shared}
                <button
                    type="button"
                    on:click={publishShoppingList}
                    class="flex flex-row items-center bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm gap-x-1"
                >
                    {#if publishing}
                        <ProgressWheel className="w-6 animate-spin" />
                        <p>Publishing...</p>
                    {:else}
                        <PublishIcon className="w-6" />
                        <p>Publish</p>
                    {/if}
                </button>
            {:else}
                <div class="inline-flex gap-1 items-center">
                    <button
                        type="button"
                        on:click={syncShoppingList}
                        class="flex flex-row items-center bg-transparent hover:bg-gray-300 transition-colors p-1 rounded-sm"
                    >
                        {#if synchronizingList}
                            <SyncIcon className="w-6 animate-spin" />
                        {:else}
                            <DownloadIcon className="w-6" />
                        {/if}
                    </button>

                    <button
                        type="button"
                        on:click={uploadShoppingList}
                        class="flex flex-row items-center bg-transparent transition-colors p-1 rounded-sm hover:bg-gray-300 disabled:opacity-50 disabled:bg-gray-300"
                    >
                        {#if uploadingList}
                            <SyncIcon className="w-6 animate-spin" />
                        {:else}
                            <UploadIcon className="w-6" />
                        {/if}
                    </button>
                </div>
            {/if}
        </div>
    </div>
    <div class="h-fit w-full mt-52">
        <div class="w-[36rem] mx-auto">
            <input
                type="text"
                name="ListName"
                id="listName"
                bind:value={data.list_info.title}
                maxlength="24"
                on:keyup={() =>
                    typewatch(() => {
                        updateListTitle();
                    }, 1000)}
                class="text-5xl hidden-placeholder focus-visible:outline-none"
            />
            {#if data.list_info.list_id}
                <span class="inline-flex gap-2">
                    <h4 class="text-sm text-slate-700 pl-1">
                        {data.list_info.list_id}
                    </h4>

                    <button
                        on:click={() => {
                            shareIdElement.select();
                            document.execCommand("copy");
                        }}
                    >
                        <ShareIcon className="w-4" />
                    </button>
                    <input
                        type="text"
                        class="w-1 opacity-0"
                        bind:this={shareIdElement}
                        value={data.list_info.list_id}
                        disabled
                    />
                </span>
            {/if}
        </div>
        <br />
        <ul class="flex flex-col gap-y-1">
            {#each data.items as item}
                <ListItem
                    bind:item
                    on:update={async () =>
                        (item = await updateItemCounter(item))}
                />
            {/each}
            <button
                type="button"
                class="w-full cursor-text"
                on:click={selectNextItem}
            >
                <textarea
                    bind:this={nextItem}
                    bind:value={nextItemValue}
                    on:input={() => {
                        nextItem.style.height = "1px";
                        nextItem.style.height = `${nextItem.scrollHeight}px`;
                    }}
                    on:keypress={(event) => {
                        if (event.code == "Enter") {
                            event.preventDefault();
                            addNewItem();
                        }
                    }}
                    name="newItem"
                    id="newItem"
                    class="text-lg p-1 w-[36rem] pl-2 hidden-placeholder focus-visible:outline-none resize-none"
                    placeholder="Input new item name"
                />
            </button>
        </ul>
    </div>
    <button
        class="h-full w-full min-h-[30vh] cursor-text"
        on:click={selectNextItem}
    />
</div>
