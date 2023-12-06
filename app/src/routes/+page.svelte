<script lang="ts">
    import { goto } from "$app/navigation";
    import ListCard from "$lib/components/ListCard.svelte";
    import AddIcon from "$lib/icons/AddIcon.svelte";
    import type { ListInfo } from "$lib/types";
    import { invoke } from "@tauri-apps/api/tauri";

    type ListPageData = {
        lists: ListInfo[];
    };

    export let data: ListPageData;

    const lists = data.lists;

    let joinListId = "";

    let joining = false;

    const createList = async () => {
        await invoke("create_list").then((value) => {
            goto("/list?id=" + value);
        });
    };

    const joinList = async () => {
        if (joining) return;

        joining = true;
        invoke("join_list", {
            listId: joinListId,
        })
            .then((value) => {
                goto("/list?id=" + value);
            })
            .catch((reason) => {
                console.log("failed to join list", reason);
            })
            .finally(() => {
                joining = false;
            });
    };
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<section class="flex flex-col justify-center items-center">
    <div class="text-5xl m-2 mt-10">Quantum List</div>
    <span class="h-[10vh]"></span>
    <div class="mt-5 inline-flex justify-between w-full max-w-6xl px-4 xl:p-0">
        <button
            type="button"
            class="inline-flex bg-sunglow hover:bg-orange-300 p-2 px-3 rounded-md"
            on:click={createList}
        >
            <AddIcon className="w-6" />
            <p>Create List</p>
        </button>
        <form on:submit={joinList}>
            <input
                type="text"
                name="listId"
                id="joinListId"
                bind:value={joinListId}
                placeholder="Paste list code here"
                class="h-full pl-2 border-2 rounded-l-md bg-white border-black"
            />
            <button
                type="submit"
                disabled={joining}
                class="inline-flex bg-sunglow hover:bg-orange-300 disabled:bg-orange-100 p-2 px-3 rounded-r-md"
                >Join</button
            >
        </form>
    </div>
    <div
        class="p-4 max-w-6xl w-full grid grid-cols-2 gap-4 sm:grid-cols-3 xl:px-0 lg:grid-cols-4"
    >
        {#each lists as list}
            <ListCard
                title={list.title}
                list_id={list.list_id}
                shared={list.shared}
            />
        {/each}
    </div>
</section>
