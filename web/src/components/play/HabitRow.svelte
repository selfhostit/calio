<script lang="ts">
  import { doneCount, scheduleLabel, type Check, type Habit } from '../../lib/play/habits';
  import Icon from '../Icon.svelte';

  // A habit for today: tap the circle to tick it off (once per time it's
  // due), anywhere else to change it.
  let {
    habit,
    day,
    checks,
    due,
    onopen,
    ontick,
  }: {
    habit: Habit;
    day: string;
    checks: Check[];
    due: boolean;
    onopen: () => void;
    ontick: () => void;
  } = $props();

  const count = $derived(doneCount(habit, day, checks));
  const done = $derived(count >= habit.times);
  const skipped = $derived(checks.some((c) => c.habit === habit.id && c.day === day && c.skipped));
</script>

<li class="flex items-center gap-3 py-2.5 {due ? '' : 'opacity-60'}">
  <button class="flex min-w-0 flex-1 items-center gap-3 text-left" onclick={onopen}>
    <span
      class="grid size-10 shrink-0 place-items-center rounded-xl {done
        ? 'bg-amber-100 text-amber-700 dark:bg-amber-400/20 dark:text-amber-200'
        : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300'}"
      ><Icon name={habit.icon} class="size-5" /></span
    >
    <span class="min-w-0">
      <span class="block truncate font-semibold {done ? 'text-zinc-500 line-through' : ''}"
        >{habit.title}</span
      >
      <span class="block truncate text-xs text-zinc-500 dark:text-zinc-400"
        >{skipped ? 'Skipped today' : scheduleLabel(habit)}</span
      >
    </span>
  </button>
  {#if due}
    {#if habit.times > 1}
      <span class="flex gap-1" aria-hidden="true">
        {#each Array.from({ length: habit.times }, (_, i) => i) as i (i)}
          <span
            class="size-1.5 rounded-full {i < count
              ? 'bg-amber-500'
              : 'bg-zinc-300 dark:bg-zinc-700'}"
          ></span>
        {/each}
      </span>
    {/if}
    <button
      class="grid size-11 shrink-0 place-items-center rounded-full border-2 transition active:scale-90 {done
        ? 'border-amber-500 bg-amber-500 text-white'
        : 'border-zinc-300 text-transparent hover:border-amber-400 dark:border-zinc-600'}"
      onclick={ontick}
      aria-label={done
        ? `${habit.title}: done, tap to start over`
        : `Tick off ${habit.title}${habit.times > 1 ? ` (${count} of ${habit.times})` : ''}`}
      aria-pressed={done}
    >
      <Icon name="check" class="size-5" />
    </button>
  {/if}
</li>
