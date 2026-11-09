<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { addDays, dayLabel, inDays } from '../lib/days';
  import { agenda, occurrences, span, whenLabel, type Entry } from '../lib/entries';
  import EntryRow from './EntryRow.svelte';
  import Icon from './Icon.svelte';

  let {
    entries,
    onopen,
    onadd,
    onimport,
  }: {
    entries: Entry[];
    onopen: (e: Entry) => void;
    onadd: () => void;
    onimport: () => void;
  } = $props();

  // Everything coming up. Repeats show for the next two weeks, or just the
  // next one when that's further out (a birthday in May shows in January).
  const groups = $derived.by(() => {
    const today = calendar.today;
    const once = entries.filter((e) => !e.repeat);
    const repeating = entries.filter((e) => e.repeat);
    const soon = occurrences(repeating, today, addDays(today, 13));
    const seen = new Set(soon.map((e) => e.id));
    const later: Entry[] = [];
    for (const e of occurrences(
      repeating.filter((e) => !seen.has(e.id)),
      addDays(today, 14),
      addDays(today, 400),
    )) {
      if (!seen.has(e.id)) (seen.add(e.id), later.push(e));
    }
    return agenda([...once, ...soon, ...later], today);
  });

  // Sitting (pets, kids, the house) coming up in the next two weeks that nobody has booked yet.
  const unbooked = $derived(
    calendar.entries
      .filter((e) => e.kind === 'petsit' && !e.repeat && e.sitter_status !== 'booked')
      .filter((e) => {
        const [first, last] = span(e);
        return last >= calendar.today && first <= addDays(calendar.today, 14);
      })
      .sort((a, b) => (a.start < b.start ? -1 : 1)),
  );

  async function book(e: Entry) {
    await calendar.save(e.id, { sitter_status: 'booked' });
  }
</script>

{#if unbooked.length}
  <section class="flex flex-col gap-2">
    {#each unbooked as e (e.id)}
      <div
        class="flex items-center gap-3 rounded-2xl border border-amber-200 bg-amber-50/70 p-3 dark:border-amber-400/20 dark:bg-amber-400/5"
      >
        <span
          class="grid size-10 shrink-0 place-items-center rounded-xl bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300"
        >
          <Icon name={calendar.look(e).icon} />
        </span>
        <button class="min-w-0 flex-1 text-left" onclick={() => onopen(e)}>
          <span class="block font-semibold text-zinc-900 dark:text-zinc-100"
            >Sitter needed: {e.title}</span
          >
          <span class="block text-sm text-zinc-600 dark:text-zinc-400"
            >{whenLabel(e)} · starts {inDays(span(e)[0], calendar.today)}{e.sitter_status ===
              'asked' && e.sitter
              ? ` · asked ${e.sitter}`
              : ''}</span
          >
        </button>
        <button class="btn-secondary shrink-0 px-3 py-2 text-sm" onclick={() => book(e)}
          >Booked</button
        >
      </div>
    {/each}
  </section>
{/if}

{#if groups.length}
  <div class="flex flex-col gap-5">
    {#each groups as group (group.day)}
      <section>
        <h3
          class="sticky top-0 z-10 -mx-1 bg-zinc-50/90 px-1 py-2 text-xs font-bold tracking-wider text-zinc-500 uppercase backdrop-blur dark:bg-zinc-950/90 dark:text-zinc-400"
        >
          {dayLabel(group.day, calendar.today)}
        </h3>
        <div class="flex flex-col">
          {#each group.entries as entry (entry.key ?? entry.id)}
            <EntryRow {entry} {onopen} />
          {/each}
        </div>
      </section>
    {/each}
  </div>
{:else if calendar.loaded}
  <div class="flex flex-col items-center gap-3 px-6 py-16 text-center">
    <div class="flex gap-2">
      <span
        class="grid size-12 -rotate-6 place-items-center rounded-2xl bg-indigo-50 text-indigo-600 dark:bg-indigo-400/10 dark:text-indigo-300"
        ><Icon name="sparkle" /></span
      >
      <span
        class="grid size-12 place-items-center rounded-2xl bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-300"
        ><Icon name="paw" /></span
      >
      <span
        class="grid size-12 rotate-6 place-items-center rounded-2xl bg-teal-50 text-teal-700 dark:bg-teal-400/10 dark:text-teal-300"
        ><Icon name="plane" /></span
      >
    </div>
    <p class="mt-2 text-lg font-semibold">Nothing coming up</p>
    <p class="text-sm text-zinc-500">
      Add a dinner, a trip, or the days someone needs to look after the pets or the kids.
    </p>
    <button class="btn-primary mt-2" onclick={onadd}>Add something</button>
    <button class="text-sm font-medium text-brand-600 dark:text-brand-400" onclick={onimport}
      >or import from Google, Outlook or Apple Calendar</button
    >
  </div>
{/if}
