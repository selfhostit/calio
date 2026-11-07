<script lang="ts">
  import { untrack } from 'svelte';
  import { calendar, firstName } from '../lib/calendar.svelte';
  import { birthdayCategory, isBirthday, offered } from '../lib/categories';
  import {
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
  import { errorMessage } from '../lib/pb';
  import { sendToShoppa, shoppaEnabled } from '../lib/shoppa';
  import { session } from '../lib/session.svelte';
  import Avatar from './Avatar.svelte';
  import Checklist from './Checklist.svelte';
  import Comments from './Comments.svelte';
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

  // the categories switched on, and this entry's own
  const choices = $derived(offered(calendar.categories, editing?.category));
  let category = $state(editing?.category ?? untrack(() => choices[0]?.id ?? ''));
  const chosen = $derived(calendar.categories.find((c) => c.id === category));
  const kind = $derived<Kind>(chosen?.kind ?? editing?.kind ?? 'event');
  const look = $derived(calendar.look({ kind, category }));
  let title = $state(editing?.title ?? '');
  let allDay = $state(editing ? editing.all_day : untrack(() => kind !== 'event'));
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
  // the sitters the household has had before, to pick from
  const sitters = [
    ...new Set(calendar.entries.map((e) => e.sitter?.trim()).filter((s): s is string => !!s)),
  ].sort();
  // where "the pets need a sitter" goes
  const petsCategory = $derived(
    choices.find((c) => c.kind === 'petsit' && c.icon === 'paw') ??
      choices.find((c) => c.kind === 'petsit'),
  );
  let repeat = $state<Repeat | ''>(editing?.repeat ?? '');
  let repeatUntil = $state(editing?.repeat_until ? editing.repeat_until.slice(0, 10) : '');
  let checklist = $state<ChecklistItem[]>(editing?.checklist ?? []);
  let showChecklist = $state(!!editing?.checklist?.length);
  let shopping = $state<ChecklistItem[]>(editing?.shopping ?? []);
  let showShopping = $state(!!editing?.shopping?.length);
  let shoppa = $state(false);
  let shoppaBusy = $state(false);
  let shoppaMessage = $state('');
  shoppaEnabled().then((on) => (shoppa = on));
  const toBuy = $derived(shopping.filter((i) => !i.done).length);

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

  // On an entry that already exists, ticking items saves at once, so the
  // lists work while you pack or shop.
  async function saveLists() {
    if (!editing) return;
    try {
      await calendar.save(editing.id, { checklist, shopping });
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function toShoppa() {
    if (!editing) return;
    shoppaBusy = true;
    shoppaMessage = '';
    try {
      const r = await sendToShoppa(editing.id);
      shoppaMessage = `Sent ${r.sent} ${r.sent === 1 ? 'item' : 'items'} to ${r.list} in Shoppa.`;
    } catch (err) {
      shoppaMessage = errorMessage(err);
    } finally {
      shoppaBusy = false;
    }
  }
  let busy = $state(false);
  let error = $state('');
  // the calendar it was subscribed from, which updates it
  const source = $derived(
    editing?.calendar ? calendar.subscriptions.find((s) => s.id === editing.calendar) : undefined,
  );

  const canDelete = $derived(
    !!editing && (editing.created_by === session.user?.id || session.isOwner),
  );

  const placeholder = $derived(
    kind === 'trip'
      ? 'Where to? Crete'
      : kind === 'event'
        ? 'Dinner at Maria’s'
        : look.icon === 'baby'
          ? 'Eva and Leo'
          : look.icon === 'home'
            ? 'The house and the plants'
            : 'Luna and Milo',
  );

  // A new entry titled "... birthday" goes in the birthday category, all day
  // and every year, until someone picks a category themselves.
  let picked = $state(false);
  let auto = false;
  const birthdays = $derived(birthdayCategory(calendar.categories));
  $effect(() => {
    if (editing || picked || !birthdays) return;
    const on = isBirthday(title);
    untrack(() => {
      if (on && !auto) {
        auto = true;
        category = birthdays.id;
        allDay = true;
        if (!repeat) repeat = 'yearly';
      } else if (!on && auto) {
        auto = false;
        category = choices[0]?.id ?? '';
        allDay = (choices[0]?.kind ?? 'event') !== 'event';
        if (repeat === 'yearly') repeat = '';
      }
    });
  });

  function pick(id: string) {
    picked = true;
    category = id;
    // trips and sitting are usually whole days
    if (!editing) allDay = kind !== 'event';
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
      category,
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
      shopping,
    };
    try {
      await calendar.save(editing?.id ?? null, data);
      if (!editing && kind === 'trip' && needSitter) {
        await calendar.save(null, {
          title: `Pets while we're in ${data.title}`,
          category: petsCategory?.id ?? '',
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
    {#if source}
      <p
        class="flex items-start gap-2 rounded-2xl bg-zinc-50 p-3 text-sm text-zinc-600 dark:bg-zinc-800/50 dark:text-zinc-300"
      >
        <Icon name="repeat" class="mt-0.5 size-4 shrink-0" />
        <span
          >From <b>{source.name}</b>. Changes there replace the title, times, place and notes here;
          what you add stays.</span
        >
      </p>
    {/if}
    {#if choices.length > 1}
      <div
        class="grid gap-2 {choices.length === 2
          ? 'grid-cols-2'
          : choices.length === 4
            ? 'grid-cols-2 sm:grid-cols-4'
            : 'grid-cols-3'}"
        role="radiogroup"
        aria-label="Category"
      >
        {#each choices as c (c.id)}
          {@const l = calendar.look({ kind: c.kind, category: c.id })}
          <button
            type="button"
            role="radio"
            aria-checked={category === c.id}
            class="flex flex-col items-center gap-1 rounded-2xl border-2 px-2 py-3 text-sm font-semibold transition {category ===
            c.id
              ? l.chip
              : 'border-zinc-200 text-zinc-600 hover:border-zinc-300 dark:border-zinc-700 dark:text-zinc-300'}"
            onclick={() => pick(c.id)}
          >
            <Icon name={l.icon} class="size-6" />
            <span class="max-w-full truncate">{c.name}</span>
          </button>
        {/each}
      </div>
    {/if}

    <label>
      {kind === 'trip' ? 'Destination' : kind === 'petsit' ? 'Who needs looking after' : 'What'}
      <!-- svelte-ignore a11y_autofocus -->
      <input bind:value={title} {placeholder} required maxlength="200" autofocus={!editing} />
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
        <input
          bind:value={sitter}
          placeholder="Name, or leave empty for now"
          maxlength="120"
          list="sitters"
        />
        <datalist id="sitters">
          {#each sitters as name (name)}
            <option value={name}></option>
          {/each}
        </datalist>
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

    {#if kind === 'trip' && !editing && petsCategory}
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
        onchange={saveLists}
      />
    {/if}

    {#if showShopping}
      <div class="flex flex-col gap-2">
        <Checklist
          bind:items={shopping}
          label="Shopping list"
          placeholder="Add something to buy"
          onchange={saveLists}
        />
        {#if shoppa && editing && toBuy}
          <button
            type="button"
            class="btn-secondary gap-1.5 self-start px-3 py-2 text-sm"
            onclick={toShoppa}
            disabled={shoppaBusy}
            ><Icon name="send" class="size-4" />{shoppaBusy
              ? 'Sending…'
              : `Send ${toBuy} to Shoppa`}</button
          >
        {:else if shoppa && !editing && shopping.length}
          <p class="text-xs text-zinc-500">Once it's added, you can send this list to Shoppa.</p>
        {/if}
        {#if shoppaMessage}<p class="text-sm text-zinc-500" role="status">{shoppaMessage}</p>{/if}
      </div>
    {/if}

    {#if !(kind === 'trip' || showChecklist) || !showShopping}
      <div class="flex flex-wrap gap-x-4 gap-y-2">
        {#if !(kind === 'trip' || showChecklist)}
          <button
            type="button"
            class="inline-flex items-center gap-1.5 text-sm font-medium text-brand-600 dark:text-brand-400"
            onclick={() => (showChecklist = true)}
            ><Icon name="check" class="size-4" /> Add a checklist</button
          >
        {/if}
        {#if !showShopping}
          <button
            type="button"
            class="inline-flex items-center gap-1.5 text-sm font-medium text-brand-600 dark:text-brand-400"
            onclick={() => (showShopping = true)}
            ><Icon name="cart" class="size-4" /> Add a shopping list</button
          >
        {/if}
      </div>
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

  {#if editing}
    <Comments entry={editing} />
  {/if}
</Sheet>
