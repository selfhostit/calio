<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { offered } from '../lib/categories';
  import {
    GUIDES,
    importFile,
    importSummary,
    looksLikeLink,
    refresh,
    subscribe,
    syncedLabel,
  } from '../lib/imports';
  import { errorMessage } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Icon from './Icon.svelte';

  // Bringing in other calendars: by their link, kept up to date, or from a
  // .ics file, copied once. With steps for the apps people use most.
  let mode = $state<'' | 'link' | 'file'>('');
  let url = $state('');
  let name = $state('');
  let file = $state<File | null>(null);
  const choices = $derived(offered(calendar.categories));
  let category = $state('');
  let busy = $state(false);
  let message = $state('');
  let refreshing = $state('');

  function open(m: 'link' | 'file') {
    mode = mode === m ? '' : m;
    message = '';
    if (!category) category = choices.find((c) => c.kind === 'event')?.id ?? choices[0]?.id ?? '';
  }

  async function addLink(e: SubmitEvent) {
    e.preventDefault();
    if (!looksLikeLink(url)) {
      message = 'Paste the whole link. It starts with https:// or webcal://.';
      return;
    }
    busy = true;
    message = '';
    try {
      const s = await subscribe(url.trim(), name.trim(), category);
      calendar.putSubscription(s);
      if (s.error) {
        message = `Added, but it couldn't be read yet: ${s.error}. Calio tries again every half hour.`;
      } else {
        message = `Added ${s.name} with ${s.events} ${s.events === 1 ? 'event' : 'events'}.`;
        url = '';
        name = '';
        mode = '';
      }
    } catch (err) {
      message = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function addFile(e: SubmitEvent) {
    e.preventDefault();
    if (!file) return;
    busy = true;
    message = '';
    try {
      message = importSummary(await importFile(file, category));
      file = null;
      mode = '';
    } catch (err) {
      message = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function update(id: string) {
    refreshing = id;
    try {
      calendar.putSubscription(await refresh(id));
    } catch (err) {
      message = errorMessage(err);
    } finally {
      refreshing = '';
    }
  }

  async function remove(id: string, label: string) {
    if (!confirm(`Remove ${label}? Its events leave the calendar too.`)) return;
    try {
      await calendar.removeSubscription(id);
    } catch (err) {
      message = errorMessage(err);
    }
  }
</script>

<section class="card flex flex-col gap-3" id="import">
  <h3 class="font-semibold">Import calendars</h3>
  <p class="text-sm text-zinc-600 dark:text-zinc-400">
    Bring in events from Google Calendar, Outlook, Apple Calendar and others. Add a calendar's link
    to keep it up to date, or import an .ics file once.
  </p>

  {#if calendar.subscriptions.length}
    <ul class="flex flex-col divide-y divide-zinc-100 dark:divide-zinc-800">
      {#each calendar.subscriptions as s (s.id)}
        {@const look = calendar.look({ kind: 'event', category: s.category })}
        <li class="flex items-center gap-3 py-2">
          <span class="grid size-8 shrink-0 place-items-center rounded-full {look.chip}"
            ><Icon name="repeat" class="size-4" /></span
          >
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold">{s.name}</p>
            {#if s.error}
              <p class="text-xs text-red-600 dark:text-red-400">{s.error}</p>
            {:else}
              <p class="text-xs text-zinc-500">
                {s.events}
                {s.events === 1 ? 'event' : 'events'} · updated {syncedLabel(s.synced)}
              </p>
            {/if}
          </div>
          <button
            class="grid size-8 place-items-center rounded-full text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
            onclick={() => update(s.id)}
            disabled={refreshing === s.id}
            aria-label="Update {s.name} now"
            ><Icon
              name="repeat"
              class="size-4 {refreshing === s.id ? 'animate-spin' : ''}"
            /></button
          >
          {#if s.created_by === session.user?.id || session.isOwner}
            <button
              class="grid size-8 place-items-center rounded-full text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
              onclick={() => remove(s.id, s.name)}
              aria-label="Remove {s.name}"><Icon name="trash" class="size-4" /></button
            >
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  <div class="flex flex-wrap gap-2">
    <button
      class="{mode === 'link' ? 'btn-primary' : 'btn-secondary'} px-3 py-2 text-sm"
      onclick={() => open('link')}
      aria-expanded={mode === 'link'}>Add a calendar link</button
    >
    <button
      class="{mode === 'file' ? 'btn-primary' : 'btn-secondary'} px-3 py-2 text-sm"
      onclick={() => open('file')}
      aria-expanded={mode === 'file'}>Import a file</button
    >
  </div>

  {#snippet categoryPicker()}
    {#if choices.length > 1}
      <label
        >Put its events in
        <select bind:value={category}>
          {#each choices as c (c.id)}<option value={c.id}>{c.name}</option>{/each}
        </select>
      </label>
    {/if}
  {/snippet}

  {#if mode === 'link'}
    <form class="flex flex-col gap-3" onsubmit={addLink}>
      <label
        >Calendar link
        <input
          bind:value={url}
          type="url"
          inputmode="url"
          required
          placeholder="https://calendar.google.com/calendar/ical/…/basic.ics"
          autocomplete="off"
        />
      </label>
      <label
        >Name <input
          bind:value={name}
          maxlength="80"
          placeholder="Optional, like Work or School holidays"
        /></label
      >
      {@render categoryPicker()}
      <p class="text-xs text-zinc-500">
        Calio checks it every half hour. Changes there replace the title, times, place and notes
        here; sitters, lists and comments you add stay.
      </p>
      <button class="btn-primary self-start" disabled={busy}
        >{busy ? 'Fetching…' : 'Add calendar'}</button
      >
    </form>
  {:else if mode === 'file'}
    <form class="flex flex-col gap-3" onsubmit={addFile}>
      <label
        >Calendar file (.ics)
        <input
          type="file"
          accept=".ics,.ical,.ifb,.icalendar,text/calendar"
          required
          onchange={(e) => (file = e.currentTarget.files?.[0] ?? null)}
        />
      </label>
      {@render categoryPicker()}
      <p class="text-xs text-zinc-500">
        A one-time copy. Importing the same file again updates its events instead of adding them
        twice.
      </p>
      <button class="btn-primary self-start" disabled={busy || !file}
        >{busy ? 'Importing…' : 'Import'}</button
      >
    </form>
  {/if}

  {#if message}<p class="text-sm text-zinc-600 dark:text-zinc-300" role="status">{message}</p>{/if}

  <div class="flex flex-col gap-1">
    <p class="text-sm font-medium">Where do I find the link or file?</p>
    {#each GUIDES as g (g.app)}
      <details class="group rounded-xl bg-zinc-50 px-3 py-2 text-sm dark:bg-zinc-800/50">
        <summary class="cursor-pointer font-medium select-none">{g.app}</summary>
        <div class="flex flex-col gap-2 pt-2 pb-1 text-zinc-600 dark:text-zinc-300">
          {#if g.link}
            <p class="font-semibold text-zinc-800 dark:text-zinc-100">
              Link <span class="font-normal text-zinc-500">(stays up to date)</span>
            </p>
            <ol class="list-decimal space-y-1 pl-5">
              {#each g.link as step (step)}<li>{step}</li>{/each}
            </ol>
          {/if}
          {#if g.file}
            <p class="font-semibold text-zinc-800 dark:text-zinc-100">
              File <span class="font-normal text-zinc-500">(copied once)</span>
            </p>
            <ol class="list-decimal space-y-1 pl-5">
              {#each g.file as step (step)}<li>{step}</li>{/each}
            </ol>
          {/if}
          {#if g.note}<p class="text-xs text-zinc-500">{g.note}</p>{/if}
        </div>
      </details>
    {/each}
  </div>
</section>
