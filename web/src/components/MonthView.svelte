<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { dateOf, daysBetween, longDayLabel, monthGrid, monthName, type Day } from '../lib/days';
  import { byStart, isOn, occurrences, span, type Entry } from '../lib/entries';
  import { moveBy } from '../lib/reschedule';
  import EntryRow from './EntryRow.svelte';
  import Icon from './Icon.svelte';

  let {
    entries,
    selected = $bindable(),
    onopen,
    onadd,
  }: {
    entries: Entry[];
    selected: Day;
    onopen: (e: Entry) => void;
    onadd: (day: Day) => void;
  } = $props();

  let year = $state(dateOf(selected).getFullYear());
  let month = $state(dateOf(selected).getMonth());

  const days = $derived(monthGrid(year, month));
  const shown = $derived(occurrences(entries, days[0], days[days.length - 1]));
  const weeks = $derived(Array.from({ length: 6 }, (_, i) => days.slice(i * 7, i * 7 + 7)));
  const LANES = 3;

  // Each week puts its entries in lanes, so a trip stays one straight bar
  // across the days it covers.
  const layout = $derived(
    weeks.map((week) => {
      const [from, to] = [week[0], week[6]];
      const inWeek = shown
        .filter((e) => {
          const [first, last] = span(e);
          return first <= to && last >= from;
        })
        .sort((a, b) => {
          const [fa, la] = span(a);
          const [fb, lb] = span(b);
          const sa = fa < from ? from : fa;
          const sb = fb < from ? from : fb;
          if (sa !== sb) return sa < sb ? -1 : 1;
          return lb.localeCompare(la) || byStart(a, b);
        });
      const lanes: (Entry | null)[][] = week.map(() => Array(LANES).fill(null));
      const more = week.map(() => 0);
      for (const e of inWeek) {
        const cols = week.map((d, i) => (isOn(e, d) ? i : -1)).filter((i) => i >= 0);
        const lane = [...Array(LANES).keys()].find((l) => cols.every((c) => !lanes[c][l]));
        if (lane === undefined) cols.forEach((c) => more[c]++);
        else cols.forEach((c) => (lanes[c][lane] = e));
      }
      return { lanes, more };
    }),
  );

  const dayEntries = $derived(
    occurrences(entries, selected, selected)
      .filter((e) => isOn(e, selected))
      .sort(byStart),
  );

  function move(n: number) {
    const d = new Date(year, month + n, 1);
    year = d.getFullYear();
    month = d.getMonth();
  }

  function goToday() {
    const d = dateOf(calendar.today);
    year = d.getFullYear();
    month = d.getMonth();
    selected = calendar.today;
  }

  // With a mouse, an entry's bar can be dragged to another day. It moves by
  // as many days as it was dragged (a repeat moves its whole series), and the
  // toast that follows can undo it. Phones move entries in the form instead.
  let dragging = $state<{ id: string; from: Day } | null>(null);
  let over = $state<Day | null>(null);

  function dragStart(ev: DragEvent, e: Entry, from: Day) {
    if (!ev.dataTransfer) return;
    ev.dataTransfer.effectAllowed = 'move';
    ev.dataTransfer.setData('text/plain', e.title);
    dragging = { id: e.id, from };
  }

  function dragOver(ev: DragEvent, day: Day) {
    if (!dragging) return;
    ev.preventDefault();
    if (ev.dataTransfer) ev.dataTransfer.dropEffect = 'move';
    over = day;
  }

  function drop(ev: DragEvent, day: Day) {
    if (!dragging) return;
    ev.preventDefault();
    const { id, from } = dragging;
    dragging = over = null;
    selected = day;
    const entry = calendar.entries.find((e) => e.id === id);
    if (entry) moveBy(entry, daysBetween(from, day));
  }

  const WEEKDAYS = Array.from({ length: 7 }, (_, i) =>
    new Intl.DateTimeFormat(undefined, { weekday: 'narrow' }).format(new Date(2024, 0, 1 + i)),
  );
</script>

<section class="flex flex-col gap-3">
  <div class="flex items-center gap-2">
    <h2 class="flex-1 text-xl font-bold tracking-tight">
      {monthName.format(new Date(year, month, 1))}
    </h2>
    <button class="btn-secondary px-3 py-2 text-sm" onclick={goToday}>Today</button>
    <button class="btn-secondary size-10 p-0" onclick={() => move(-1)} aria-label="Previous month"
      ><Icon name="left" /></button
    >
    <button class="btn-secondary size-10 p-0" onclick={() => move(1)} aria-label="Next month"
      ><Icon name="right" /></button
    >
  </div>

  <div
    class="overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
  >
    <div
      class="grid grid-cols-7 border-b border-zinc-200 text-center text-xs font-semibold text-zinc-400 dark:border-zinc-800"
    >
      {#each WEEKDAYS as w, i (i)}<div class="py-2">{w}</div>{/each}
    </div>
    {#each weeks as week, w (week[0])}
      <div
        class="grid grid-cols-7 border-b border-zinc-100 last:border-b-0 dark:border-zinc-800/70"
      >
        {#each week as day, i (day)}
          {@const inMonth = dateOf(day).getMonth() === month}
          {@const isToday = day === calendar.today}
          {@const isSelected = day === selected}
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="relative flex min-h-16 flex-col gap-0.5 py-1 sm:min-h-24 {i > 0
              ? 'border-l border-zinc-100 dark:border-zinc-800/70'
              : ''}"
            ondragover={(ev) => dragOver(ev, day)}
            ondragleave={() => over === day && (over = null)}
            ondrop={(ev) => drop(ev, day)}
          >
            <!-- the whole day picks it; the bars on top of it can be dragged -->
            <button
              class="absolute inset-0 transition {isSelected
                ? 'bg-zinc-100 dark:bg-zinc-800/70'
                : 'hover:bg-zinc-50 dark:hover:bg-zinc-800/50'} {over === day
                ? 'bg-brand-500/10! ring-2 ring-brand-500 ring-inset'
                : ''}"
              onclick={() => (selected = day)}
              ondblclick={() => onadd(day)}
              aria-label={longDayLabel(day)}
              aria-pressed={isSelected}
            ></button>
            <span
              class="pointer-events-none relative mx-auto grid size-7 place-items-center rounded-full text-sm sm:mx-1.5 {isToday
                ? 'bg-brand-600 font-semibold text-white'
                : inMonth
                  ? 'font-medium'
                  : 'text-zinc-300 dark:text-zinc-600'}">{dateOf(day).getDate()}</span
            >
            {#each layout[w].lanes[i] as e, lane (lane)}
              {#if e}
                {@const [first, last] = span(e)}
                {@const starts = first === day || i === 0}
                {@const ends = last === day || i === 6}
                {@const look = calendar.look(e)}
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <span
                  class="pointer-events-none relative block h-1.5 truncate px-1.5 text-[11px] leading-5 font-medium select-none sm:h-5 pointer-fine:pointer-events-auto pointer-fine:cursor-grab {look.pill} {look.bar} {starts
                    ? 'ml-1 rounded-l-full'
                    : '-ml-px'} {ends ? 'mr-1 rounded-r-full' : ''} {inMonth
                    ? ''
                    : 'opacity-40'} {dragging?.id === e.id ? 'opacity-50' : ''}"
                  draggable="true"
                  title="{e.title}: drag to move"
                  ondragstart={(ev) => dragStart(ev, e, day)}
                  ondragend={() => (dragging = over = null)}
                  onclick={() => onopen(e)}
                >
                  <span class="hidden sm:inline">{starts ? e.title : ' '}</span>
                </span>
              {:else}
                <span class="block h-1.5 sm:h-5"></span>
              {/if}
            {/each}
            {#if layout[w].more[i]}
              <span
                class="pointer-events-none relative px-1.5 text-[10px] font-semibold text-zinc-400"
                >+{layout[w].more[i]}</span
              >
            {/if}
          </div>
        {/each}
      </div>
    {/each}
  </div>

  <div class="mt-2 flex items-center justify-between">
    <h3 class="font-semibold">{longDayLabel(selected)}</h3>
    <button
      class="inline-flex items-center gap-1 text-sm font-semibold text-brand-600 dark:text-brand-400"
      onclick={() => onadd(selected)}><Icon name="plus" class="size-4" /> Add</button
    >
  </div>
  {#if dayEntries.length}
    <div class="-mx-3 flex flex-col">
      {#each dayEntries as entry (entry.key ?? entry.id)}<EntryRow {entry} {onopen} />{/each}
    </div>
  {:else}
    <p class="text-sm text-zinc-500">Nothing on this day.</p>
  {/if}
</section>
