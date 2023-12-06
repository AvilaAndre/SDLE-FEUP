<script lang="ts">
    import { goto } from "$app/navigation";
    import type { TabInfo } from "$lib/types";
    import { closeTab, openTab } from "$lib/writables/listTabs";

    export let tab: TabInfo;

    function openThisTab() {
        openTab(tab.title, tab.ref);
        goto(tab.ref);
    }
</script>

<li
    class={"inline-flex max-w-[12rem] px-1 py-1 h-fit items-end group rounded-t-md" +
        " " +
        (tab.selected ? "tab-selected bg-white" : "")}
>
    <button
        type="button"
        on:click={openThisTab}
        class="w-44 text-left hover:bg-gray-100 rounded-md inline-flex gap-2 p-1 px-2 h-fit group-[.tab-selected]:hover:bg-white"
    >
        <p class="w-full whitespace-nowrap text-ellipsis overflow-hidden">
            {tab.title}
        </p>
        {#if tab.deletable}
            <button
                type="button"
                class="group-hover:block hidden text-gray-700 hover:text-black w-6 h-6"
                on:click={() => closeTab(tab.ref)}
            >
                X
            </button>
        {/if}
    </button>
</li>
