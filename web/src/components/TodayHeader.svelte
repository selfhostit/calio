<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { longDayLabel } from '../lib/days';
  import { isOn, occurrences, progressLabel } from '../lib/entries';
  import { sheets } from '../lib/sheets.svelte';
  import Icon from './Icon.svelte';

  // What today looks like, above every section: the date, a one-line summary
  // and what's on.
  const today = $derived(
    occurrences(calendar.entries, calendar.today, calendar.today).filter((e) =>
      isOn(e, calendar.today),
    ),
  );
  const away = $derived(today.find((e) => e.kind === 'trip'));

  const summary = $derived(
    away
      ? away.title
      : today.length === 0
        ? 'Nothing planned today'
        : today.length === 1
          ? `${today[0].title} today`
          : `${today.length} things today`,
  );
</script>

<p class="mt-6 text-sm font-medium text-zinc-500 dark:text-zinc-400">
  {longDayLabel(calendar.today)}{#if away}&nbsp;· away, {progressLabel(
      away,
      calendar.today,
    ).toLowerCase() || 'today'}{/if}
</p>
<h1 class="mt-0.5 text-3xl font-bold tracking-tight">{summary}</h1>
{#if today.length}
  <div class="mt-3 flex flex-wrap gap-1.5">
    {#each today.slice(0, 4) as e (e.key ?? e.id)}
      {@const look = calendar.look(e)}
      <button
        class="inline-flex max-w-full items-center gap-1.5 rounded-full border border-zinc-200 bg-white py-1 pr-3 pl-1.5 text-sm font-medium hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900 dark:hover:bg-zinc-800"
        onclick={() => sheets.open(e)}
      >
        <span class="grid size-5 place-items-center rounded-full {look.tile}"
          ><Icon name={look.icon} class="size-3.5" /></span
        >
        <span class="truncate">{e.title}</span>
      </button>
    {/each}
  </div>
{/if}
