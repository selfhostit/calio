<script lang="ts">
  import { onMount } from 'svelte';
  import { calendar } from '../lib/calendar.svelte';
  import {
    KINDS,
    longDayLabel,
    isOn,
    progressLabel,
    type Day,
    type Entry,
    type Kind,
  } from '../lib/entries';
  import { LOOK } from '../lib/look';
  import { session } from '../lib/session.svelte';
  import Agenda from './Agenda.svelte';
  import Avatar from './Avatar.svelte';
  import EntrySheet from './EntrySheet.svelte';
  import Icon from './Icon.svelte';
  import MonthView from './MonthView.svelte';
  import Settings from './Settings.svelte';

  type View = 'upcoming' | 'month';
  const saved = (() => {
    try {
      return localStorage.getItem('calio.view') as View | null;
    } catch {
      return null;
    }
  })();

  let view = $state<View>(saved === 'month' ? 'month' : 'upcoming');
  let filter = $state<Kind | 'all'>('all');
  let selected = $state<Day>(calendar.today);
  let sheet = $state<{ entry: Entry | null; day: Day } | null>(null);
  let settings = $state(false);

  const entries = $derived(
    filter === 'all' ? calendar.entries : calendar.entries.filter((e) => e.kind === filter),
  );
  const today = $derived(calendar.entries.filter((e) => isOn(e, calendar.today)));
  const away = $derived(today.find((e) => e.kind === 'trip'));

  $effect(() => {
    try {
      localStorage.setItem('calio.view', view);
    } catch {
      // private mode: the view just isn't remembered
    }
  });

  onMount(() => {
    calendar.start();
    return () => calendar.stop();
  });

  const open = (entry: Entry) => (sheet = { entry, day: calendar.today });
  const add = (day: Day = view === 'month' ? selected : calendar.today) =>
    (sheet = { entry: null, day });

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

<div class="mx-auto flex min-h-dvh w-full max-w-3xl flex-col">
  <header
    class="relative overflow-hidden px-4 pt-[max(1rem,env(safe-area-inset-top))] pb-6 text-white"
  >
    <div
      class="absolute inset-0 -z-10 bg-gradient-to-br from-violet-600 via-fuchsia-500 to-orange-400 sm:rounded-b-3xl"
    ></div>
    <div class="absolute -top-16 -right-10 -z-10 size-56 rounded-full bg-white/10 blur-2xl"></div>
    <div class="flex items-center justify-between">
      <span class="inline-flex items-center gap-2 text-lg font-extrabold tracking-tight">
        <img src="/icons/icon.svg" alt="" class="size-7 rounded-lg ring-1 ring-white/30" /> Calio
      </span>
      <button class="rounded-full" onclick={() => (settings = true)} aria-label="Settings">
        <Avatar person={session.user!} size="size-9 text-sm" />
      </button>
    </div>
    <p class="mt-6 text-sm font-medium text-white/80">
      {longDayLabel(calendar.today)}{#if away}&nbsp;· away, {progressLabel(
          away,
          calendar.today,
        ).toLowerCase() || 'today'}{/if}
    </p>
    <h1 class="mt-0.5 text-3xl font-extrabold tracking-tight">{summary}</h1>
    {#if today.length}
      <div class="mt-3 flex flex-wrap gap-1.5">
        {#each today.slice(0, 4) as e (e.id)}
          <button
            class="inline-flex max-w-full items-center gap-1.5 rounded-full bg-white/20 py-1 pr-3 pl-1.5 text-sm font-medium backdrop-blur hover:bg-white/30"
            onclick={() => open(e)}
          >
            <span
              class="grid size-5 place-items-center rounded-full bg-white/90 {LOOK[e.kind].tile
                .split(' ')
                .find((c) => c.startsWith('text-'))}"
              ><Icon name={LOOK[e.kind].icon} class="size-3.5" /></span
            >
            <span class="truncate">{e.title}</span>
          </button>
        {/each}
      </div>
    {/if}
  </header>

  <main class="flex flex-1 flex-col gap-4 px-4 pt-4 pb-28">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="inline-flex rounded-xl bg-zinc-200/70 p-1 dark:bg-zinc-800" role="tablist">
        {#each [{ v: 'upcoming', label: 'Upcoming', icon: 'list' }, { v: 'month', label: 'Month', icon: 'grid' }] as t (t.v)}
          <button
            role="tab"
            aria-selected={view === t.v}
            class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-semibold transition {view ===
            t.v
              ? 'bg-white shadow-sm dark:bg-zinc-700'
              : 'text-zinc-500'}"
            onclick={() => (view = t.v as View)}
            ><Icon name={t.icon} class="size-4" />{t.label}</button
          >
        {/each}
      </div>
      <div class="-mx-4 flex gap-1.5 overflow-x-auto px-4 sm:mx-0 sm:px-0">
        <button
          class="shrink-0 rounded-full border px-3 py-1.5 text-sm font-medium transition {filter ===
          'all'
            ? 'border-zinc-900 bg-zinc-900 text-white dark:border-white dark:bg-white dark:text-zinc-900'
            : 'border-zinc-300 text-zinc-600 dark:border-zinc-700 dark:text-zinc-300'}"
          onclick={() => (filter = 'all')}>All</button
        >
        {#each KINDS as k (k.kind)}
          <button
            class="inline-flex shrink-0 items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition {filter ===
            k.kind
              ? LOOK[k.kind].chip
              : 'border-zinc-300 text-zinc-600 dark:border-zinc-700 dark:text-zinc-300'}"
            onclick={() => (filter = filter === k.kind ? 'all' : k.kind)}
          >
            <span class="size-2 rounded-full {filter === k.kind ? 'bg-white' : LOOK[k.kind].dot}"
            ></span>
            {k.plural}
          </button>
        {/each}
      </div>
    </div>

    {#if calendar.error}
      <div class="card flex items-center gap-3 text-sm" role="alert">
        <Icon name="alert" class="size-5 text-red-500" />
        <p class="flex-1">{calendar.error}</p>
        <button
          class="btn-secondary px-3 py-2 text-sm"
          onclick={() => ((calendar.error = ''), calendar.start())}>Try again</button
        >
      </div>
    {:else if !calendar.loaded}
      <div class="flex flex-col gap-3" aria-busy="true">
        {#each [0, 1, 2] as i (i)}
          <div class="h-16 animate-pulse rounded-2xl bg-zinc-200/70 dark:bg-zinc-800"></div>
        {/each}
      </div>
    {:else if view === 'upcoming'}
      <Agenda {entries} onopen={open} onadd={() => add()} />
    {:else}
      <MonthView {entries} bind:selected onopen={open} onadd={add} />
    {/if}
  </main>

  <button
    class="fixed right-5 bottom-[max(1.25rem,env(safe-area-inset-bottom))] z-30 grid size-16 place-items-center rounded-2xl bg-gradient-to-br from-violet-600 to-fuchsia-500 text-white shadow-xl shadow-violet-500/30 transition hover:scale-105 active:scale-95 sm:right-[max(1.25rem,calc(50vw-24rem+1.25rem))]"
    onclick={() => add()}
    aria-label="Add to the calendar"
  >
    <Icon name="plus" class="size-7" />
  </button>
</div>

{#if sheet}
  <EntrySheet entry={sheet.entry} day={sheet.day} onclose={() => (sheet = null)} />
{/if}
{#if settings}
  <Settings onclose={() => (settings = false)} />
{/if}
