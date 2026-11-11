<script lang="ts">
  import { onMount } from 'svelte';
  import { calendar } from '../lib/calendar.svelte';
  import { play } from '../lib/play/play.svelte';
  import { session } from '../lib/session.svelte';
  import { sheets } from '../lib/sheets.svelte';
  import Avatar from './Avatar.svelte';
  import EntrySheet from './EntrySheet.svelte';
  import Icon from './Icon.svelte';
  import CalendarSection from './sections/CalendarSection.svelte';
  import CompanionSection from './sections/CompanionSection.svelte';
  import DocumentsSection from './sections/DocumentsSection.svelte';
  import Settings from './Settings.svelte';
  import Toast from './Toast.svelte';
  import TodayHeader from './TodayHeader.svelte';

  // The signed-in app: the brand and settings at the top, today's summary,
  // and one section at a time, picked with the tabs. A section brings its own
  // toolbar, content and add button; the tabs are passed in for its toolbar.

  type View = 'upcoming' | 'month' | 'documents' | 'companion';
  const TABS: { v: View; label: string; icon: string }[] = [
    { v: 'upcoming', label: 'Upcoming', icon: 'list' },
    { v: 'month', label: 'Month', icon: 'grid' },
    { v: 'documents', label: 'Docs', icon: 'doc' },
    { v: 'companion', label: 'Panda', icon: 'paw' },
  ];

  const saved = (() => {
    try {
      return localStorage.getItem('calio.view') as View | null;
    } catch {
      return null;
    }
  })();
  let view = $state<View>(TABS.some((t) => t.v === saved) ? saved! : 'upcoming');

  $effect(() => {
    try {
      localStorage.setItem('calio.view', view);
    } catch {
      // private mode: the view just isn't remembered
    }
  });

  onMount(() => {
    calendar.start();
    return () => {
      calendar.stop();
      play.stop();
    };
  });
</script>

{#snippet tabs()}
  <div class="inline-flex rounded-xl bg-zinc-200/70 p-1 dark:bg-zinc-800" role="tablist">
    {#each TABS as t (t.v)}
      <button
        role="tab"
        aria-selected={view === t.v}
        class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-semibold transition {view ===
        t.v
          ? 'bg-white shadow-sm dark:bg-zinc-700'
          : 'text-zinc-500'}"
        onclick={() => (view = t.v)}
        ><Icon name={t.icon} class="size-4" /><span class={view === t.v ? '' : 'max-sm:sr-only'}
          >{t.label}</span
        ></button
      >
    {/each}
  </div>
{/snippet}

<div class="mx-auto flex min-h-dvh w-full max-w-3xl flex-col">
  <header class="px-4 pt-[max(1rem,env(safe-area-inset-top))] pb-2">
    <div class="flex items-center justify-between">
      <span class="inline-flex items-center gap-2 text-lg font-bold tracking-tight">
        <img src="/icons/icon.svg" alt="" class="size-7 rounded-lg" /> Calio
      </span>
      <button class="rounded-full" onclick={() => (sheets.settings = 'open')} aria-label="Settings">
        <Avatar person={session.user!} size="size-9 text-sm" />
      </button>
    </div>
    <TodayHeader />
  </header>

  <main class="flex flex-1 flex-col gap-4 px-4 pt-4 pb-28">
    {#if view === 'companion'}
      <CompanionSection {tabs} />
    {:else if view === 'documents'}
      <DocumentsSection {tabs} />
    {:else}
      <CalendarSection mode={view} {tabs} />
    {/if}
  </main>
</div>

{#if sheets.entry}
  <EntrySheet
    entry={sheets.entry.entry}
    day={sheets.entry.day}
    onclose={() => (sheets.entry = null)}
  />
{/if}
{#if sheets.settings}
  <Settings importing={sheets.settings === 'import'} onclose={() => (sheets.settings = false)} />
{/if}
<Toast />
