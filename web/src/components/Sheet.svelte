<script lang="ts">
  import type { Snippet } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import Icon from './Icon.svelte';

  // A panel that slides up from the bottom on phones and sits in the middle on
  // larger screens. Escape or a tap outside closes it.
  let { title, onclose, children }: { title: string; onclose: () => void; children: Snippet } =
    $props();

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose();
  }
</script>

<svelte:window {onkeydown} />

<div class="fixed inset-0 z-40 flex items-end justify-center sm:items-center sm:p-6">
  <button
    class="absolute inset-0 bg-zinc-950/40 backdrop-blur-[2px]"
    transition:fade={{ duration: 150 }}
    onclick={onclose}
    aria-label="Close"
    tabindex="-1"
  ></button>
  <div
    class="relative flex max-h-[92dvh] w-full max-w-lg flex-col overflow-hidden rounded-t-3xl bg-white shadow-2xl sm:rounded-3xl dark:bg-zinc-900"
    role="dialog"
    aria-modal="true"
    aria-label={title}
    transition:fly={{ y: 40, duration: 220 }}
  >
    <div class="mx-auto mt-2 h-1 w-10 rounded-full bg-zinc-300 sm:hidden dark:bg-zinc-700"></div>
    <header class="flex items-center justify-between px-5 pt-3 pb-2">
      <h2 class="text-lg font-bold">{title}</h2>
      <button
        class="grid size-9 place-items-center rounded-full bg-zinc-100 text-zinc-600 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-300"
        onclick={onclose}
        aria-label="Close"><Icon name="close" class="size-4" /></button
      >
    </header>
    <div class="overflow-y-auto px-5 pb-[max(1.25rem,env(safe-area-inset-bottom))]">
      {@render children()}
    </div>
  </div>
</div>
