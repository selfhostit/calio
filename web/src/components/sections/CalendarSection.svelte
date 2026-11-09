<script lang="ts">
  import type { Snippet } from 'svelte';
  import { calendar } from '../../lib/calendar.svelte';
  import { offered } from '../../lib/categories';
  import type { Day } from '../../lib/days';
  import { sheets } from '../../lib/sheets.svelte';
  import Agenda from '../Agenda.svelte';
  import Fab from '../Fab.svelte';
  import Loaded from '../Loaded.svelte';
  import MonthView from '../MonthView.svelte';

  // The household calendar, as a list of what's coming up or as a month.
  let { mode, tabs }: { mode: 'upcoming' | 'month'; tabs: Snippet } = $props();

  // 'all', or a category's id
  let filter = $state('all');
  const chips = $derived(offered(calendar.categories));
  let selected = $state<Day>(calendar.today);

  const entries = $derived(
    filter === 'all' ? calendar.entries : calendar.entries.filter((e) => e.category === filter),
  );

  // a filter for a category that was switched off or deleted goes away
  $effect(() => {
    if (filter !== 'all' && !chips.some((c) => c.id === filter)) filter = 'all';
  });

  const add = (day: Day = mode === 'month' ? selected : calendar.today) => sheets.add(day);
</script>

<div class="flex flex-wrap items-center justify-between gap-3">
  {@render tabs()}
  <div
    class="-mx-4 flex gap-1.5 overflow-x-auto px-4 sm:mx-0 sm:px-0 {chips.length < 2
      ? 'hidden'
      : ''}"
  >
    <button
      class="shrink-0 rounded-full border px-3 py-1.5 text-sm font-medium transition {filter ===
      'all'
        ? 'border-zinc-300 bg-zinc-100 text-zinc-900 dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-100'
        : 'border-zinc-300 text-zinc-600 dark:border-zinc-700 dark:text-zinc-300'}"
      onclick={() => (filter = 'all')}>All</button
    >
    {#each chips as c (c.id)}
      {@const look = calendar.look({ kind: c.kind, category: c.id })}
      <button
        class="inline-flex shrink-0 items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition {filter ===
        c.id
          ? look.chip
          : 'border-zinc-300 text-zinc-600 dark:border-zinc-700 dark:text-zinc-300'}"
        onclick={() => (filter = filter === c.id ? 'all' : c.id)}
      >
        <span class="size-2 rounded-full {look.dot}"></span>
        {c.name}
      </button>
    {/each}
  </div>
</div>

<Loaded
  loaded={calendar.loaded}
  error={calendar.error}
  onretry={() => ((calendar.error = ''), calendar.start())}
>
  {#if mode === 'upcoming'}
    <Agenda
      {entries}
      onopen={(e) => sheets.open(e)}
      onadd={() => add()}
      onimport={() => (sheets.settings = 'import')}
    />
  {:else}
    <MonthView {entries} bind:selected onopen={(e) => sheets.open(e)} onadd={add} />
  {/if}
</Loaded>

<Fab label="Add to the calendar" onclick={() => add()} />
