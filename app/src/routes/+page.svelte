<script lang="ts">
    import { goto } from "$app/navigation";
    import ListCard from "$lib/components/ListCard.svelte";
    import AddIcon from "$lib/icons/AddIcon.svelte";
    import { invoke } from "@tauri-apps/api/tauri";

    type ListInfo = {
        title: string;
        id: string;
        shared: boolean;
    };

    type ListPageData = {
        lists: ListInfo[];
    };

    export let data: ListPageData;

    const lists = data.lists;

    const createList = async () => {
        await invoke("create_list").then((value) => {
            goto("/list?id=" + value);
        });
    };
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<section class="flex flex-col justify-center items-center">
    <div class="text-2xl m-2">Quantum List</div>
    <div>
        <button
            type="button"
            class="inline-flex bg-sunglow p-1 rounded-sm"
            on:click={createList}
        >
            <AddIcon className="w-6" />
            <p>Create List</p>
        </button>
    </div>
    <div class="grid grid-cols-3 gap-4">
        {#each lists as list}
            <ListCard title={list.title} id={list.id} shared={list.shared} />
        {/each}
    </div>
</section>
