<script lang="ts">
    export let title: string;
    export let id: string;
</script>

<a
    href={"/list?id=" + id}
    class="bg-white p-2 h-32 w-64 shadow-md rounded-md disabled card-1"
>
    <h4 class="text-xl">{title}</h4>
    <h4 class="text-xs text-slate-500">{id}</h4>
</a>
