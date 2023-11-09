<script lang="ts">
    type ListPageData = {
        title: string;
        id: string;
        items: string[];
    };

    export let data: ListPageData;
</script>

<svelte:head>
    <title>Home</title>
    <meta name="description" content="App" />
</svelte:head>

<div class="bg-red-200 w-72">
    <ul>
        <li>
            <a href="/"> main page </a>
        </li>
    </ul>
</div>
<div class="flex flex-col justify-start items-center w-full">
    <div class="h-full w-[36rem] mt-12">
        <h1 class="text-4xl">{data.title}</h1>
        <h4 class="text-sm text-slate-700">{data.id}</h4>
        <br />
        <ul>
            {#each data.items as item}
                <li class="text-lg pl-2 hover:bg-white">{item}</li>
            {/each}
        </ul>
    </div>
</div>
