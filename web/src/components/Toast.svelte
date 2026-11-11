<script lang="ts">
  import { fly } from 'svelte/transition';
  import { toasts } from '../lib/toast.svelte';
  import Icon from './Icon.svelte';

  // The toast from lib/toast.svelte.ts, above the add button. Screen readers
  // hear it as it comes.
  // a list of at most one, so a toast that is fading out keeps its own text
  const shown = $derived(toasts.current ? [toasts.current] : []);
</script>

<div
  class="pointer-events-none fixed inset-x-0 bottom-[max(5.5rem,calc(env(safe-area-inset-bottom)+4.25rem))] z-50 flex justify-center px-4"
  role="status"
  aria-live="polite"
>
  {#each shown as toast (toast.id)}
    <div
      class="pointer-events-auto flex max-w-md items-center gap-3 rounded-2xl bg-zinc-900 py-2.5 pr-2 pl-4 text-sm text-white shadow-xl shadow-zinc-900/20 dark:bg-zinc-100 dark:text-zinc-900"
      transition:fly={{ y: 16, duration: 180 }}
    >
      <span class="min-w-0 flex-1">{toast.text}</span>
      {#if toast.action}
        {@const action = toast.action}
        <button
          class="inline-flex shrink-0 items-center gap-1 rounded-xl px-3 py-1.5 font-semibold text-brand-400 hover:bg-white/10 dark:text-brand-700 dark:hover:bg-zinc-900/10"
          onclick={() => (toasts.dismiss(toast.id), action.run())}
          ><Icon name="undo" class="size-4" />{action.label}</button
        >
      {/if}
      <button
        class="grid size-8 shrink-0 place-items-center rounded-full text-white/60 hover:bg-white/10 dark:text-zinc-500 dark:hover:bg-zinc-900/10"
        onclick={() => toasts.dismiss(toast.id)}
        aria-label="Dismiss"><Icon name="close" class="size-4" /></button
      >
    </div>
  {/each}
</div>
