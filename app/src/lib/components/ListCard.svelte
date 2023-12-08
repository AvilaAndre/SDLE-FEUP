<script lang="ts">
    import PublishIcon from "$lib/icons/PublishIcon.svelte";
    import SavedLocallyIcon from "$lib/icons/SavedLocallyIcon.svelte";

    export let title: string;
    export let list_id: string;
    export let shared: boolean | undefined = false;
</script>

<a
    href={"/list?id=" + list_id}
    class="bg-white p-2 h-fit w-full shadow-md rounded-md disabled card-1 relative"
>
    <div>
        <span class="w-full inline-flex justify-between gap-x-1">
            <h3
                class="text-xl overflow-x-hidden overflow-ellipsis whitespace-nowrap"
            >
                {title}
            </h3>
            <span>
                {#if shared}
                    <PublishIcon className="w-6" />
                {:else}
                    <SavedLocallyIcon className="w-6" />
                {/if}
            </span>
        </span>
        <h4 class="text-xs text-slate-500">{list_id}</h4>
    </div>
</a>
