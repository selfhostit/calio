<script lang="ts">
  import { onMount, type Snippet } from 'svelte';
  import { errorMessage } from '../../lib/pb';
  import { PARTS, doneCount, isDue, type Habit } from '../../lib/play/habits';
  import { play } from '../../lib/play/play.svelte';
  import { today } from '../../lib/today.svelte';
  import Fab from '../Fab.svelte';
  import Loaded from '../Loaded.svelte';
  import CompanionCard from '../play/CompanionCard.svelte';
  import HabitRow from '../play/HabitRow.svelte';
  import HabitSheet from '../play/HabitSheet.svelte';
  import MeetCompanion from '../play/MeetCompanion.svelte';

  // Your red panda and today's habits.
  let { tabs }: { tabs: Snippet } = $props();

  let sheet = $state<{ habit: Habit | null } | null>(null);
  let error = $state('');

  onMount(() => {
    play.start();
  });

  const due = $derived(play.active.filter((h) => isDue(h, today.day, play.checks)));
  const later = $derived(play.active.filter((h) => !isDue(h, today.day, play.checks)));
  const done = $derived(due.filter((h) => doneCount(h, today.day, play.checks) >= h.times));
  const groups = $derived(
    PARTS.map((p) => ({ ...p, habits: due.filter((h) => h.part === p.part) })).filter(
      (g) => g.habits.length,
    ),
  );

  async function tick(h: Habit) {
    error = '';
    try {
      await play.tick(h);
    } catch (err) {
      error = errorMessage(err, "Couldn't save that. Please try again.");
    }
  }
</script>

<div class="flex flex-wrap items-center justify-between gap-3">
  {@render tabs()}
</div>

<Loaded loaded={play.loaded} error={play.error} onretry={() => ((play.error = ''), play.start())}>
  {#if !play.companion}
    <MeetCompanion />
  {:else}
    <CompanionCard
      companion={play.companion}
      mood={play.mood}
      line={play.line(due.length > 0 && done.length === due.length)}
      done={done.length}
      due={due.length}
    />

    {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}

    {#each groups as g (g.part)}
      <section class="card py-2">
        <h3 class="pt-2 text-xs font-semibold tracking-wide text-zinc-500 uppercase">
          {g.label}
        </h3>
        <ul class="divide-y divide-zinc-100 dark:divide-zinc-800">
          {#each g.habits as h (h.id)}
            <HabitRow
              habit={h}
              day={today.day}
              checks={play.checks}
              due
              onopen={() => (sheet = { habit: h })}
              ontick={() => tick(h)}
            />
          {/each}
        </ul>
      </section>
    {:else}
      {#if !later.length}
        <div class="card flex flex-col items-center gap-3 py-8 text-center">
          <p class="font-semibold">No habits yet</p>
          <p class="text-sm text-zinc-500">Add one small thing to do each day.</p>
          <button class="btn-primary" onclick={() => (sheet = { habit: null })}>Add a habit</button>
        </div>
      {/if}
    {/each}

    {#if later.length}
      <section class="card py-2">
        <h3 class="pt-2 text-xs font-semibold tracking-wide text-zinc-500 uppercase">Not today</h3>
        <ul class="divide-y divide-zinc-100 dark:divide-zinc-800">
          {#each later as h (h.id)}
            <HabitRow
              habit={h}
              day={today.day}
              checks={play.checks}
              due={false}
              onopen={() => (sheet = { habit: h })}
              ontick={() => {}}
            />
          {/each}
        </ul>
      </section>
    {/if}
  {/if}
</Loaded>

{#if play.companion}
  <Fab label="Add a habit" onclick={() => (sheet = { habit: null })} />
{/if}

{#if sheet}
  <HabitSheet habit={sheet.habit} onclose={() => (sheet = null)} />
{/if}
