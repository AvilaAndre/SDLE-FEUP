<script>
    import Tabs from "$lib/components/Tabs.svelte";
    import "../app.css";

    import { tabsList } from "$lib/writables/listTabs";
</script>

<div class="app">
    <section class="bg-gray-200 w-screen h-10 fixed">
        <Tabs bind:tabs={$tabsList} />
    </section>
    <main>
        <slot />
    </main>
</div>
