<script lang="ts">
  import { today } from '../../lib/today.svelte';
  import { errorMessage } from '../../lib/pb';
  import {
    HABIT_ICONS,
    PARTS,
    WEEKDAYS,
    checkOf,
    type Habit,
    type Part,
    type Weekday,
  } from '../../lib/play/habits';
  import { play } from '../../lib/play/play.svelte';
  import Icon from '../Icon.svelte';
  import Sheet from '../Sheet.svelte';

  // Adds a habit, or changes one: what, how often, when. An existing one can
  // be skipped today or put away.
  let { habit, onclose }: { habit: Habit | null; onclose: () => void } = $props();

  // svelte-ignore state_referenced_locally
  const h = habit;
  let title = $state(h?.title ?? '');
  let icon = $state(h?.icon ?? 'heart');
  let times = $state(h?.times ?? 1);
  let days = $state<Weekday[]>(h?.days?.length ? [...h.days] : WEEKDAYS.map((w) => w.day));
  let part = $state<Part | ''>(h?.part ?? '');
  let busy = $state(false);
  let error = $state('');

  const skipped = $derived(h ? !!checkOf(play.checks, h, today.day)?.skipped : false);

  function toggleDay(d: Weekday) {
    days = days.includes(d) ? days.filter((x) => x !== d) : [...days, d];
  }

  async function run(fn: () => Promise<unknown>) {
    busy = true;
    error = '';
    try {
      await fn();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function save(e: SubmitEvent) {
    e.preventDefault();
    if (!days.length) return (error = 'Pick at least one day.');
    run(() =>
      play.saveHabit(h?.id ?? null, {
        title: title.trim(),
        icon,
        times,
        // every day is stored as none, so it stays every day
        days: days.length === 7 ? [] : WEEKDAYS.map((w) => w.day).filter((d) => days.includes(d)),
        part,
      }),
    );
  }
</script>

<Sheet title={h ? 'Change habit' : 'New habit'} {onclose}>
  <form class="flex flex-col gap-4 pt-1" onsubmit={save}>
    <label>
      What
      <input bind:value={title} required maxlength="60" placeholder="Brush teeth" />
    </label>

    <div class="flex flex-wrap gap-1.5" role="radiogroup" aria-label="Icon">
      {#each HABIT_ICONS as x (x)}
        <button
          type="button"
          role="radio"
          aria-checked={icon === x}
          aria-label={x}
          class="grid size-10 place-items-center rounded-xl border transition {icon === x
            ? 'border-amber-400 bg-amber-100 text-amber-800 dark:border-amber-500/60 dark:bg-amber-400/20 dark:text-amber-200'
            : 'border-zinc-200 text-zinc-500 dark:border-zinc-700'}"
          onclick={() => (icon = x)}><Icon name={x} class="size-5" /></button
        >
      {/each}
    </div>

    <fieldset class="flex flex-col gap-1.5">
      <legend class="mb-1.5 text-sm font-medium">On</legend>
      <div class="flex gap-1">
        {#each WEEKDAYS as w (w.day)}
          <button
            type="button"
            aria-pressed={days.includes(w.day)}
            class="h-10 flex-1 rounded-xl border text-sm font-semibold transition {days.includes(
              w.day,
            )
              ? 'border-amber-400 bg-amber-100 text-amber-800 dark:border-amber-500/60 dark:bg-amber-400/20 dark:text-amber-200'
              : 'border-zinc-200 text-zinc-500 dark:border-zinc-700'}"
            onclick={() => toggleDay(w.day)}>{w.label}</button
          >
        {/each}
      </div>
    </fieldset>

    <div class="grid grid-cols-2 gap-3">
      <label>
        Times a day
        <select bind:value={times}>
          {#each [1, 2, 3, 4, 5] as n (n)}
            <option value={n}>{n === 1 ? 'Once' : n === 2 ? 'Twice' : `${n} times`}</option>
          {/each}
        </select>
      </label>
      <label>
        When
        <select bind:value={part}>
          {#each PARTS as p (p.part)}
            <option value={p.part}>{p.label}</option>
          {/each}
        </select>
      </label>
    </div>

    {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}

    <button class="btn-primary" disabled={busy}>{h ? 'Save' : 'Add habit'}</button>
    {#if h}
      <div class="grid grid-cols-2 gap-3">
        <button
          type="button"
          class="btn-secondary gap-2"
          disabled={busy}
          onclick={() => run(() => play.skip(h, !skipped))}
          ><Icon name="skip" class="size-4" />{skipped ? 'Not skipping' : 'Skip today'}</button
        >
        <button
          type="button"
          class="btn-secondary gap-2 text-red-600 dark:text-red-400"
          disabled={busy}
          onclick={() => run(() => play.archive(h))}
          ><Icon name="trash" class="size-4" />Remove</button
        >
      </div>
    {/if}
  </form>
</Sheet>
