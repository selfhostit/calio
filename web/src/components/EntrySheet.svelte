<script lang="ts">
  import { untrack } from 'svelte';
  import { calendar, firstName } from '../lib/calendar.svelte';
  import {
    KINDS,
    SITTER_STATUS,
    allDayValue,
    instantValue,
    localInputValue,
    span,
    type Day,
    type Entry,
    type Kind,
    type SitterStatus,
    type ChecklistItem,
    type Repeat,
    REPEATS,
    mapLink,
  } from '../lib/entries';
  import { LOOK } from '../lib/look';
  import { errorMessage } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Avatar from './Avatar.svelte';
  import Checklist from './Checklist.svelte';
  import Icon from './Icon.svelte';
  import Sheet from './Sheet.svelte';

  let {
    entry = null,
    day,
    onclose,
  }: { entry?: Entry | null; day: Day; onclose: () => void } = $props();

  // the form starts from the entry (or day) it was opened with
  const { editing, first } = untrack(() => ({ editing: entry, first: day }));
  const initial = editing ? span(editing) : [first, first];

  let kind = $state<Kind>(editing?.kind ?? 'event');
  let title = $state(editing?.title ?? '');
  let allDay = $state(editing ? editing.all_day : false);
  let startDay = $state(initial[0]);
  let endDay = $state(initial[1]);
  let startAt = $state(
    editing && !editing.all_day ? localInputValue(editing.start) : `${first}T18:00`,
  );
  let endAt = $state(editing && !editing.all_day ? localInputValue(editing.end) : '');
  let location = $state(editing?.location ?? '');
  let notes = $state(editing?.notes ?? '');
  let people = $state<string[]>(editing?.people ?? (session.user ? [session.user.id] : []));
  let sitter = $state(editing?.sitter ?? '');
  let sitterStatus = $state<SitterStatus>((editing?.sitter_status as SitterStatus) || 'needed');
  let needSitter = $state(false);
  let repeat = $state<Repeat | ''>(editing?.repeat ?? '');
  let repeatUntil = $state(editing?.repeat_until ? editing.repeat_until.slice(0, 10) : '');
  let checklist = $state<ChecklistItem[]>(editing?.checklist ?? []);
  let showChecklist = $state(!!editing?.checklist?.length);

  // Starting points for a packing list: the last trip's list, or a short one.
  const lastTrip = calendar.entries
    .filter((e) => e.kind === 'trip' && e.id !== editing?.id && e.checklist?.length)
    .sort((a, b) => (a.start < b.start ? 1 : -1))[0];
  const packingSuggestions = [
    ...(lastTrip
      ? [{ label: `Same as ${lastTrip.title}`, items: lastTrip.checklist!.map((i) => i.text) }]
      : []),
    {
      label: 'Starter list',
      items: [
        'Passports and IDs',
        'Tickets',
        'Phone chargers',
        'Medicine',
        'Toiletries',
        'Clothes',
        'Keys',
      ],
    },
  ];

  // On an entry that already exists, ticking items saves at once, so the list
  // works while you pack.
  async function saveChecklist() {
    if (!editing) return;
    try {
      await calendar.save(editing.id, { checklist });
    } catch (err) {
      error = errorMessage(err);
    }
  }
  let busy = $state(false);
  let error = $state('');

  const canDelete = $derived(
    !!editing && (editing.created_by === session.user?.id || session.isOwner),
  );

  const PLACEHOLDER: Record<Kind, string> = {
    event: 'Dinner at Maria’s',
    petsit: 'Who needs looking after? Luna and Milo',
    trip: 'Where to? Crete',
  };

  function pick(k: Kind) {
    kind = k;
    // trips and pet sitting are usually whole days
    if (!editing) allDay = k !== 'event';
  }

  function togglePerson(id: string) {
    people = people.includes(id) ? people.filter((p) => p !== id) : [...people, id];
  }

  $effect(() => {
    if (endDay && endDay < startDay) endDay = startDay;
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    const data: Partial<Entry> = {
      title: title.trim(),
      kind,
      all_day: allDay,
      start: allDay ? allDayValue(startDay) : instantValue(startAt),
      end: allDay
        ? endDay && endDay !== startDay
          ? allDayValue(endDay)
          : ''
        : instantValue(endAt),
      location: kind === 'petsit' ? '' : location.trim(),
      notes: notes.trim(),
      people,
      sitter: kind === 'petsit' ? sitter.trim() : '',
      sitter_status: kind === 'petsit' ? sitterStatus : '',
      repeat,
      repeat_until: repeat && repeatUntil ? allDayValue(repeatUntil) : '',
      checklist,
    };
    try {
      await calendar.save(editing?.id ?? null, data);
      if (!editing && kind === 'trip' && needSitter) {
        await calendar.save(null, {
          title: `Pets while we're in ${data.title}`,
          kind: 'petsit',
          all_day: data.all_day,
          start: data.start,
          end: data.end,
          sitter_status: 'needed',
        });
      }
      onclose();
    } catch (err) {
      const fields = (err as { response?: { data?: Record<string, { message: string }> } }).response
        ?.data;
      const first = fields && Object.values(fields)[0]?.message;
      error = first || errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove() {
    if (!editing || !confirm(`Delete “${editing.title}”?`)) return;
    busy = true;
    try {
      await calendar.remove(editing.id);
      onclose();
    } catch (err) {
      error = errorMessage(err);
      busy = false;
    }
  }
</script>

<Sheet title={editing ? 'Edit' : 'Add to the calendar'} {onclose}>
  <form class="flex flex-col gap-4 pt-1" onsubmit={save}>
    <div class="grid grid-cols-3 gap-2" role="radiogroup" aria-label="Kind">
      {#each KINDS as k (k.kind)}
        <button
          type="button"
          role="radio"
          aria-checked={kind === k.kind}
          class="flex flex-col items-center gap-1 rounded-2xl border-2 px-2 py-3 text-sm font-semibold transition {kind ===
          k.kind
            ? LOOK[k.kind].chip
            : 'border-zinc-200 text-zinc-600 hover:border-zinc-300 dark:border-zinc-700 dark:text-zinc-300'}"
          onclick={() => pick(k.kind)}
        >
          <Icon name={LOOK[k.kind].icon} class="size-6" />
          {k.label}
        </button>
      {/each}
    </div>

    <label>
      {kind === 'trip' ? 'Destination' : kind === 'petsit' ? 'Who needs looking after' : 'What'}
      <!-- svelte-ignore a11y_autofocus -->
      <input
        bind:value={title}
        placeholder={PLACEHOLDER[kind]}
        required
        maxlength="200"
        autofocus={!editing}
      />
    </label>

    <div class="flex flex-col gap-3 rounded-2xl bg-zinc-50 p-3 dark:bg-zinc-800/50">
      <label class="flex-row items-center justify-between">
        <span class="inline-flex items-center gap-2"
          ><Icon name="sun" class="size-4" /> All day</span
        >
        <input type="checkbox" bind:checked={allDay} />
      </label>
      {#if allDay}
        <div class="grid grid-cols-2 gap-2">
          <label>From <input type="date" bind:value={startDay} required /></label>
          <label>Until <input type="date" bind:value={endDay} min={startDay} /></label>
        </div>
      {:else}
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <label>Starts <input type="datetime-local" bind:value={startAt} required /></label>
          <label>Ends <input type="datetime-local" bind:value={endAt} min={startAt} /></label>
        </div>
      {/if}
      <div class="grid grid-cols-2 gap-2">
        <label
          >Repeats
          <select bind:value={repeat}>
            {#each REPEATS as r (r.repeat)}<option value={r.repeat}>{r.label}</option>{/each}
          </select>
        </label>
        {#if repeat}
          <label
            >Until <input
              type="date"
              bind:value={repeatUntil}
              min={allDay ? startDay : startAt.slice(0, 10)}
            /></label
          >
        {/if}
      </div>
    </div>

    {#if kind === 'petsit'}
      <label>
        Sitter
        <input bind:value={sitter} placeholder="Name, or leave empty for now" maxlength="120" />
      </label>
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium">Status</span>
        <div
          class="grid grid-cols-3 gap-1 rounded-xl bg-zinc-100 p-1 dark:bg-zinc-800"
          role="radiogroup"
          aria-label="Sitter status"
        >
          {#each SITTER_STATUS as s (s.status)}
            <button
              type="button"
              role="radio"
              aria-checked={sitterStatus === s.status}
              class="rounded-lg py-2 text-sm font-semibold transition {sitterStatus === s.status
                ? 'bg-white shadow-sm dark:bg-zinc-700'
                : 'text-zinc-500'}"
              onclick={() => (sitterStatus = s.status)}>{s.label}</button
            >
          {/each}
        </div>
      </div>
    {:else}
      <div class="flex flex-col gap-1.5">
        <label>
          {kind === 'trip' ? 'Staying at' : 'Where'}
          <input
            bind:value={location}
            placeholder="Address, place, “lat, lon” or a map link"
            maxlength="200"
          />
        </label>
        {#if location.trim()}
          <a
            class="inline-flex items-center gap-1 self-start text-sm font-medium text-brand-600 dark:text-brand-400"
            href={mapLink(location)}
            target="_blank"
            rel="noopener noreferrer"><Icon name="pin" class="size-4" /> Open in Google Maps</a
          >
        {/if}
      </div>
    {/if}

    {#if calendar.people.length > 1}
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium">{kind === 'trip' ? 'Who’s going' : 'Who'}</span>
        <div class="flex flex-wrap gap-2">
          {#each calendar.people as p (p.id)}
            {@const on = people.includes(p.id)}
            <button
              type="button"
              class="inline-flex items-center gap-2 rounded-full border py-1 pr-3 pl-1 text-sm font-medium transition {on
                ? 'border-brand-400 bg-brand-500/10 text-zinc-900 dark:text-zinc-100'
                : 'border-zinc-200 text-zinc-500 dark:border-zinc-700'}"
              aria-pressed={on}
              onclick={() => togglePerson(p.id)}
            >
              <Avatar person={p} size="size-6 text-[10px]" />
              {firstName(p)}
            </button>
          {/each}
        </div>
      </div>
    {/if}

    {#if kind === 'trip' && !editing}
      <label
        class="flex-row items-center gap-3 rounded-2xl border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-700 dark:bg-zinc-800/50"
      >
        <input type="checkbox" bind:checked={needSitter} />
        <span class="flex-1">
          <span class="block font-semibold">The pets need a sitter</span>
          <span class="block text-xs font-normal text-zinc-500"
            >Adds pet sitting for the same dates.</span
          >
        </span>
        <Icon name="paw" class="size-5 text-amber-600 dark:text-amber-300" />
      </label>
    {/if}

    {#if kind === 'trip' || showChecklist}
      <Checklist
        bind:items={checklist}
        label={kind === 'trip' ? 'Packing list' : 'Checklist'}
        suggestions={kind === 'trip' ? packingSuggestions : []}
        onchange={saveChecklist}
      />
    {:else}
      <button
        type="button"
        class="inline-flex items-center gap-1.5 self-start text-sm font-medium text-brand-600 dark:text-brand-400"
        onclick={() => (showChecklist = true)}
        ><Icon name="check" class="size-4" /> Add a checklist</button
      >
    {/if}

    <label>
      Notes
      <textarea bind:value={notes} rows="2" maxlength="2000" placeholder="Optional"></textarea>
    </label>

    {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}

    <div class="flex gap-2 pt-1">
      {#if canDelete}
        <button
          type="button"
          class="btn-secondary text-red-600 dark:text-red-400"
          onclick={remove}
          disabled={busy}
          aria-label="Delete"><Icon name="trash" /></button
        >
      {/if}
      <button class="btn-primary flex-1" disabled={busy}
        >{busy ? 'Saving…' : editing ? 'Save' : 'Add'}</button
      >
    </div>
  </form>
</Sheet>
