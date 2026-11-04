<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import {
    KINDS,
    REPEATS,
    checklistLabel,
    mapLink,
    placeLabel,
    progressLabel,
    whenLabel,
    type Entry,
  } from '../lib/entries';
  import { LOOK } from '../lib/look';
  import Avatar from './Avatar.svelte';
  import Icon from './Icon.svelte';

  let { entry, onopen }: { entry: Entry; onopen: (e: Entry) => void } = $props();

  const look = $derived(LOOK[entry.kind]);
  const progress = $derived(progressLabel(entry, calendar.today));
  const packed = $derived(checklistLabel(entry));
  const status = $derived(
    entry.kind === 'petsit'
      ? entry.sitter_status === 'booked'
        ? {
            text: entry.sitter ? `${entry.sitter} booked` : 'Booked',
            cls: 'bg-emerald-50 text-emerald-800 dark:bg-emerald-400/10 dark:text-emerald-300',
          }
        : entry.sitter_status === 'asked'
          ? {
              text: entry.sitter ? `Asked ${entry.sitter}` : 'Asked',
              cls: 'bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300',
            }
          : {
              text: 'Sitter needed',
              cls: 'bg-amber-50 text-amber-800 dark:bg-amber-400/10 dark:text-amber-300',
            }
      : null,
  );
</script>

<!-- The whole row opens the entry; the map link sits above it and opens the map. -->
<div
  class="relative flex w-full items-start gap-3 rounded-2xl p-3 text-left transition hover:bg-zinc-100 has-[>button:active]:scale-[0.99] dark:hover:bg-zinc-800/60"
>
  <button
    class="absolute inset-0 rounded-2xl"
    onclick={() => onopen(entry)}
    aria-label="Open {entry.title}"
  ></button>
  <span class="pointer-events-none grid size-11 shrink-0 place-items-center rounded-xl {look.tile}">
    <Icon name={look.icon} class="size-5" />
  </span>
  <span class="pointer-events-none flex min-w-0 flex-1 flex-col gap-0.5">
    <span class="flex items-center gap-2">
      <span class="truncate font-semibold">{entry.title}</span>
      {#if entry.repeat}<Icon name="repeat" class="size-3.5 shrink-0 text-zinc-400" />{/if}
      <span class="sr-only"
        >({KINDS.find((k) => k.kind === entry.kind)?.label}{entry.repeat
          ? ', ' + REPEATS.find((r) => r.repeat === entry.repeat)?.label.toLowerCase()
          : ''})</span
      >
    </span>
    <span
      class="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-zinc-500 dark:text-zinc-400"
    >
      <span>{whenLabel(entry)}</span>
      {#if entry.location}
        <a
          class="pointer-events-auto relative inline-flex min-w-0 items-center gap-0.5 underline decoration-zinc-300 underline-offset-2 hover:text-zinc-900 dark:decoration-zinc-600 dark:hover:text-zinc-100"
          href={mapLink(entry.location)}
          target="_blank"
          rel="noopener noreferrer"
          ><Icon name="pin" class="size-3.5 shrink-0" /><span class="truncate"
            >{placeLabel(entry.location)}</span
          ></a
        >
      {/if}
    </span>
    {#if status || progress || packed}
      <span class="mt-1 flex flex-wrap gap-1.5 text-xs font-medium">
        {#if progress}<span
            class="rounded-full bg-zinc-100 px-2 py-0.5 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300"
            >{progress}</span
          >{/if}
        {#if status}<span class="rounded-full px-2 py-0.5 {status.cls}">{status.text}</span>{/if}
        {#if packed}<span
            class="inline-flex items-center gap-1 rounded-full bg-zinc-100 px-2 py-0.5 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300"
            ><Icon name="check" class="size-3" />{packed}</span
          >{/if}
      </span>
    {/if}
  </span>
  {#if entry.people.length}
    <span class="pointer-events-none flex shrink-0 -space-x-2 pt-1">
      {#each entry.people.slice(0, 3) as id (id)}
        <Avatar person={calendar.person(id)} />
      {/each}
    </span>
  {/if}
</div>
