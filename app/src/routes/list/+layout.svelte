<div class="flex justify-between">
    <section class="bg-red-200 w-72 h-screen fixed">
        <ul>
            <li>
                <a href="/"> main page </a>
            </li>
        </ul>
    </section>
    <div class="w-72 flex-shrink-0" />
    <!-- Occupies the space used by the drawer when open -->

    <section class="min-h-screen mx-auto flex flex-row w-full">
        <slot />
    </section>
</div>
