<script lang="ts">
  import { onMount } from 'svelte';
  import { calendar } from '../lib/calendar.svelte';
  import { offered } from '../lib/categories';
  import type { Doc } from '../lib/documents';
  import {
    longDayLabel,
    isOn,
    occurrences,
    progressLabel,
    type Day,
    type Entry,
  } from '../lib/entries';
  import { session } from '../lib/session.svelte';
  import Agenda from './Agenda.svelte';
  import Avatar from './Avatar.svelte';
  import DocumentSheet from './DocumentSheet.svelte';
  import Documents from './Documents.svelte';
  import EntrySheet from './EntrySheet.svelte';
  import Icon from './Icon.svelte';
  import MonthView from './MonthView.svelte';
  import Settings from './Settings.svelte';

  type View = 'upcoming' | 'month' | 'documents';
  const saved = (() => {
    try {
      return localStorage.getItem('calio.view') as View | null;
    } catch {
      return null;
    }
  })();

  let view = $state<View>(saved === 'month' || saved === 'documents' ? saved : 'upcoming');
  let docSheet = $state<{ doc: Doc | null } | null>(null);
  // 'all', or a category's id
  let filter = $state('all');
  const chips = $derived(offered(calendar.categories));
  let selected = $state<Day>(calendar.today);
  let sheet = $state<{ entry: Entry | null; day: Day } | null>(null);
  let settings = $state(false);

  const entries = $derived(
    filter === 'all' ? calendar.entries : calendar.entries.filter((e) => e.category === filter),
  );
  const today = $derived(
    occurrences(calendar.entries, calendar.today, calendar.today).filter((e) =>
      isOn(e, calendar.today),
    ),
  );
  const away = $derived(today.find((e) => e.kind === 'trip'));

  // a filter for a category that was switched off or deleted goes away
  $effect(() => {
    if (filter !== 'all' && !chips.some((c) => c.id === filter)) filter = 'all';
  });

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

  // a repeat opens the entry it repeats, so edits apply to all of them
  const open = (entry: Entry) =>
    (sheet = {
      entry: calendar.entries.find((e) => e.id === entry.id) ?? entry,
      day: calendar.today,
    });
  const TABS = [
    { v: 'upcoming', label: 'Upcoming', icon: 'list' },
    { v: 'month', label: 'Month', icon: 'grid' },
    { v: 'documents', label: 'Docs', icon: 'doc' },
  ];

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
  <header class="px-4 pt-[max(1rem,env(safe-area-inset-top))] pb-2">
    <div class="flex items-center justify-between">
      <span class="inline-flex items-center gap-2 text-lg font-bold tracking-tight">
        <img src="/icons/icon.svg" alt="" class="size-7 rounded-lg" /> Calio
      </span>
      <button class="rounded-full" onclick={() => (settings = true)} aria-label="Settings">
        <Avatar person={session.user!} size="size-9 text-sm" />
      </button>
    </div>
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
            onclick={() => open(e)}
          >
            <span class="grid size-5 place-items-center rounded-full {look.tile}"
              ><Icon name={look.icon} class="size-3.5" /></span
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
        {#each TABS as t (t.v)}
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
      <div
        class="-mx-4 flex gap-1.5 overflow-x-auto px-4 sm:mx-0 sm:px-0 {view === 'documents' ||
        chips.length < 2
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
    {:else if view === 'month'}
      <MonthView {entries} bind:selected onopen={open} onadd={add} />
    {:else}
      <Documents onopen={(doc) => (docSheet = { doc })} onadd={() => (docSheet = { doc: null })} />
    {/if}
  </main>

  <button
    class="fixed right-5 bottom-[max(1.25rem,env(safe-area-inset-bottom))] z-30 grid size-14 place-items-center rounded-full bg-brand-600 text-white shadow-lg shadow-zinc-900/20 transition hover:bg-brand-700 active:scale-95 sm:right-[max(1.25rem,calc(50vw-24rem+1.25rem))]"
    onclick={() => (view === 'documents' ? (docSheet = { doc: null }) : add())}
    aria-label={view === 'documents' ? 'Add a document' : 'Add to the calendar'}
  >
    <Icon name="plus" class="size-6" />
  </button>
</div>

{#if sheet}
  <EntrySheet entry={sheet.entry} day={sheet.day} onclose={() => (sheet = null)} />
{/if}
{#if docSheet}
  <DocumentSheet doc={docSheet.doc} onclose={() => (docSheet = null)} />
{/if}
{#if settings}
  <Settings onclose={() => (settings = false)} />
{/if}
