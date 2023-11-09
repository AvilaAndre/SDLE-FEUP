<script lang="ts">
    import ListCard from "$lib/components/ListCard.svelte";
    import AddIcon from "$lib/icons/AddIcon.svelte";

    type List = {
        title: string;
        id: string;
    };

    const lists: List[] = [
        { title: "List1", id: "99e856ff-fa4e-4269-b087-fbf91a226f3c" },
        { title: "List2", id: "671f3ff3-748b-4016-afc1-b6d4a0bdfab1" },
        { title: "List3", id: "6b18023b-ed3f-4b93-808e-966b2d8dfb47" },
        { title: "List4", id: "d977fc87-e5c1-4b06-bc21-42b553ebf5d0" },
        { title: "List5", id: "bff4b478-c9f2-44db-b8fa-402302810bf9" },
        { title: "List6", id: "916859a7-596d-4192-b132-f7c37d1d8bd0" },
    ];
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<section class="flex flex-col justify-center items-center">
    <div class="text-xl m-2">Lists</div>
    <div>New List Join List</div>
    <div class="grid grid-cols-3 gap-4">
        {#each lists as list}
            <ListCard title={list.title} id={list.id} />
        {/each}
    </div>
</section>
