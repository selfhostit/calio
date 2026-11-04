<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { KINDS, progressLabel, whenLabel, type Entry } from '../lib/entries';
  import { LOOK } from '../lib/look';
  import Avatar from './Avatar.svelte';
  import Icon from './Icon.svelte';

  let { entry, onopen }: { entry: Entry; onopen: (e: Entry) => void } = $props();

  const look = $derived(LOOK[entry.kind]);
  const progress = $derived(progressLabel(entry, calendar.today));
  const status = $derived(
    entry.kind === 'petsit'
      ? entry.sitter_status === 'booked'
        ? {
            text: entry.sitter ? `${entry.sitter} booked` : 'Booked',
            cls: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300',
          }
        : entry.sitter_status === 'asked'
          ? {
              text: entry.sitter ? `Asked ${entry.sitter}` : 'Asked',
              cls: 'bg-sky-100 text-sky-800 dark:bg-sky-500/15 dark:text-sky-300',
            }
          : {
              text: 'Sitter needed',
              cls: 'bg-amber-100 text-amber-900 dark:bg-amber-500/15 dark:text-amber-300',
            }
      : null,
  );
</script>

<button
  class="group flex w-full items-start gap-3 rounded-2xl p-3 text-left transition hover:bg-zinc-100 active:scale-[0.99] dark:hover:bg-zinc-800/60"
  onclick={() => onopen(entry)}
>
  <span class="grid size-11 shrink-0 place-items-center rounded-xl {look.tile}">
    <Icon name={look.icon} class="size-5" />
  </span>
  <span class="flex min-w-0 flex-1 flex-col gap-0.5">
    <span class="flex items-center gap-2">
      <span class="truncate font-semibold">{entry.title}</span>
      <span class="sr-only">({KINDS.find((k) => k.kind === entry.kind)?.label})</span>
    </span>
    <span
      class="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-zinc-500 dark:text-zinc-400"
    >
      <span>{whenLabel(entry)}</span>
      {#if entry.location}
        <span class="inline-flex min-w-0 items-center gap-0.5"
          ><Icon name="pin" class="size-3.5 shrink-0" /><span class="truncate"
            >{entry.location}</span
          ></span
        >
      {/if}
    </span>
    {#if status || progress}
      <span class="mt-1 flex flex-wrap gap-1.5 text-xs font-medium">
        {#if progress}<span
            class="rounded-full bg-zinc-900 px-2 py-0.5 text-white dark:bg-white dark:text-zinc-900"
            >{progress}</span
          >{/if}
        {#if status}<span class="rounded-full px-2 py-0.5 {status.cls}">{status.text}</span>{/if}
      </span>
    {/if}
  </span>
  {#if entry.people.length}
    <span class="flex shrink-0 -space-x-2 pt-1">
      {#each entry.people.slice(0, 3) as id (id)}
        <Avatar person={calendar.person(id)} />
      {/each}
    </span>
  {/if}
</button>
