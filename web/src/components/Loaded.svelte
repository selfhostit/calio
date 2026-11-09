<script lang="ts">
  import type { Snippet } from 'svelte';
  import Icon from './Icon.svelte';

  // A section's content once its data is in: a skeleton while it loads, and
  // the error with a way to try again when it couldn't.
  let {
    loaded,
    error,
    onretry,
    children,
  }: { loaded: boolean; error: string; onretry: () => void; children: Snippet } = $props();
</script>

{#if error}
  <div class="card flex items-center gap-3 text-sm" role="alert">
    <Icon name="alert" class="size-5 text-red-500" />
    <p class="flex-1">{error}</p>
    <button class="btn-secondary px-3 py-2 text-sm" onclick={onretry}>Try again</button>
  </div>
{:else if !loaded}
  <div class="flex flex-col gap-3" aria-busy="true">
    {#each [0, 1, 2] as i (i)}
      <div class="h-16 animate-pulse rounded-2xl bg-zinc-200/70 dark:bg-zinc-800"></div>
    {/each}
  </div>
{:else}
  {@render children()}
{/if}
