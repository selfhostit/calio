<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { addDays, agenda, dayLabel, inDays, span, whenLabel, type Entry } from '../lib/entries';
  import EntryRow from './EntryRow.svelte';
  import Icon from './Icon.svelte';

  let {
    entries,
    onopen,
    onadd,
  }: { entries: Entry[]; onopen: (e: Entry) => void; onadd: () => void } = $props();

  const groups = $derived(agenda(entries, calendar.today));

  // Pet sitting coming up in the next two weeks that nobody has booked yet.
  const unbooked = $derived(
    calendar.entries
      .filter((e) => e.kind === 'petsit' && e.sitter_status !== 'booked')
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
        class="flex items-center gap-3 rounded-2xl border border-amber-300/70 bg-gradient-to-r from-amber-50 to-orange-50 p-3 dark:border-amber-500/30 dark:from-amber-500/10 dark:to-orange-500/10"
      >
        <span
          class="grid size-10 shrink-0 place-items-center rounded-xl bg-amber-400 text-amber-950"
        >
          <Icon name="paw" />
        </span>
        <button class="min-w-0 flex-1 text-left" onclick={() => onopen(e)}>
          <span class="block font-semibold text-amber-950 dark:text-amber-100"
            >Sitter needed: {e.title}</span
          >
          <span class="block text-sm text-amber-800 dark:text-amber-300/80"
            >{whenLabel(e)} · starts {inDays(span(e)[0], calendar.today)}{e.sitter_status ===
              'asked' && e.sitter
              ? ` · asked ${e.sitter}`
              : ''}</span
          >
        </button>
        <button
          class="shrink-0 rounded-xl bg-amber-950/90 px-3 py-2 text-sm font-semibold text-amber-50 hover:bg-amber-950 dark:bg-amber-400 dark:text-amber-950"
          onclick={() => book(e)}>Booked</button
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
          {#each group.entries as entry (entry.id)}
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
        class="grid size-12 -rotate-6 place-items-center rounded-2xl bg-violet-100 text-violet-600 dark:bg-violet-500/20 dark:text-violet-300"
        ><Icon name="sparkle" /></span
      >
      <span
        class="grid size-12 place-items-center rounded-2xl bg-amber-100 text-amber-600 dark:bg-amber-500/20 dark:text-amber-300"
        ><Icon name="paw" /></span
      >
      <span
        class="grid size-12 rotate-6 place-items-center rounded-2xl bg-sky-100 text-sky-600 dark:bg-sky-500/20 dark:text-sky-300"
        ><Icon name="plane" /></span
      >
    </div>
    <p class="mt-2 text-lg font-semibold">Nothing coming up</p>
    <p class="text-sm text-zinc-500">
      Add a dinner, a trip, or the days someone needs to look after the pets.
    </p>
    <button class="btn-primary mt-2" onclick={onadd}>Add something</button>
  </div>
{/if}
