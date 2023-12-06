<script lang="ts">
    import type { TabInfo } from "$lib/types";
    import Tab from "./Tab.svelte";

    export let tabs: TabInfo[] = [
        {
            title: "My Lists",
            ref: "/",
            deletable: false,
            selected: true,
        },
        {
            title: "Another My Lists but title longer",
            ref: "/",
            deletable: true,
            selected: false,
        },
    ];
</script>

<ul
    class="w-screen h-12 mx-auto inline-flex justify-left items-end bg-red-200 gap-x-1 overflow-scroll"
>
    {#each tabs as tab}
        <Tab bind:tab />
    {/each}
</ul>
