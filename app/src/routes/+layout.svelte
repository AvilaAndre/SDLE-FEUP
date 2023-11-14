<script>
    import Tabs from "$lib/components/Tabs.svelte";
    import "../app.css";

    import { tabsList } from "$lib/writables/listTabs";
</script>

<div class="app">
    <section class="bg-gray-200 w-screen h-12 fixed">
        <Tabs bind:tabs={$tabsList} />
    </section>
    <div class="h-12" />
    <main>
        <slot />
    </main>
</div>
