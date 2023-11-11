<script lang="ts">
    import PublishIcon from "$lib/icons/PublishIcon.svelte";
    import SavedLocallyIcon from "$lib/icons/SavedLocallyIcon.svelte";

    export let id: number;
    export let title: string;
    export let share_id: string;
    export let shared: boolean = false;
</script>

<a
    href={"/list?id=" + id}
    class="bg-white p-2 h-32 w-64 shadow-md rounded-md disabled card-1 relative"
>
    <h4 class="text-xl">{title}</h4>
    <h4 class="text-xs text-slate-500">{share_id}</h4>
    <span class="absolute right-1 bottom-1">
        {#if shared}
            <PublishIcon className="w-6" />
        {:else}
            <SavedLocallyIcon className="w-6" />
        {/if}
    </span>
</a>
