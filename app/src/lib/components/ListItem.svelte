<script lang="ts">
    import type { ListItemInfo } from "$lib/types";

    import { createEventDispatcher } from "svelte";
    import { typewatch } from "../../utils/typewatch";
    import Check from "$lib/icons/Check.svelte";

    export let item: ListItemInfo;

    const dispatch = createEventDispatcher();

    const updateCounter = () => {
        dispatch("update");
    };

    const UPDATE_TYPEWATCH_DELAY = 500;
</script>

<li class="w-full group">
    <div
        class="flex items-center justify-between text-lg w-[36rem] h-10 mx-auto group-hover:bg-gray-100 hover:cursor-pointer break-words relative"
    >
        <label
            class="flex flex-col w-full h-10 pl-2 justify-center cursor-pointer group/check"
        >
            <input
                type="checkbox"
                class="absolute hidden"
                bind:checked={item.checked}
                on:change={updateCounter}
            />
            {#if item.checked}
                <div class="absolute -left-8">
                    <Check className="w-8 fill-green-500" />
                </div>
            {:else}
                <div class="absolute -left-8">
                    <Check
                        className="w-8 fill-green-500 hidden opacity-40 group-hover/check:block"
                    />
                </div>
            {/if}
            <p class={item.checked ? "line-through" : ""}>
                {item.name}
            </p>
        </label>
        <div class="w-fit h-10 inline-flex overflow-hidden shrink-0">
            <button
                type="button"
                class="w-8 bg-gray-200 hover:bg-gray-400"
                on:click={() => {
                    item.qtd = Math.max(item.qtd - 1, 0);
                    typewatch(() => {
                        updateCounter();
                    }, UPDATE_TYPEWATCH_DELAY);
                }}>-</button
            >
            <input
                type="text"
                bind:value={item.qtd}
                on:click={(ev) => {
                    ev.target?.select();
                }}
                on:input={(input) => {
                    if (input == null) return;

                    if (input.inputType == "insertText") {
                        if (!/^[0-9]+$/.test(input.data)) {
                            let strQtd = item.qtd.toString();

                            item.qtd = parseInt(
                                strQtd.slice(0, strQtd.length - 1)
                            );
                        }
                    }

                    if (!/^[0-9]+$/.test(item.qtd.toString())) {
                        item.qtd = 0;
                    } else {
                        // Gets rid of 0s on the left
                        item.qtd = parseInt(item.qtd.toString());
                        typewatch(() => {
                            updateCounter();
                        }, UPDATE_TYPEWATCH_DELAY);
                    }
                }}
                class="w-12 bg-transparent text-center"
            />
            <button
                type="button"
                class="w-8 bg-gray-200 hover:bg-gray-400"
                on:click={() => {
                    item.qtd++;
                    typewatch(() => {
                        updateCounter();
                    }, UPDATE_TYPEWATCH_DELAY);
                }}>+</button
            >
        </div>
    </div>
</li>
