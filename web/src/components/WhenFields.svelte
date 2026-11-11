<script lang="ts">
  import type { Snippet } from 'svelte';
  import { lengthLabel, problem, shifted, withAllDay, withStart, type When } from '../lib/when';
  import Icon from './Icon.svelte';

  // When an entry happens: all day or not, its first and last day or its start
  // and end time. Changing the start moves the end along, so it keeps how long
  // it lasts; `moves` adds buttons to move it a day or a week at once. Fields
  // that belong with the dates (how it repeats) come in as children.
  let {
    when = $bindable(),
    moves = false,
    children,
  }: { when: When; moves?: boolean; children?: Snippet } = $props();

  const length = $derived(lengthLabel(when));
  const wrong = $derived(problem(when));

  const MOVES = [
    { days: -1, label: '1 day earlier', short: '−1 day' },
    { days: 1, label: '1 day later', short: '+1 day' },
    { days: 7, label: '1 week later', short: '+1 week' },
  ];

  function setEndTime(t: string) {
    when = { ...when, endTime: t, endDay: when.endTime || !t ? when.endDay : when.startDay };
  }
  function setEndDay(d: string) {
    if (!d) return;
    when = { ...when, endDay: d, endTime: when.endTime || when.startTime };
  }
</script>

<div class="flex flex-col gap-3 rounded-2xl bg-zinc-50 p-3 dark:bg-zinc-800/50">
  <label class="flex-row items-center justify-between">
    <span class="inline-flex items-center gap-2"><Icon name="sun" class="size-4" /> All day</span>
    <input
      type="checkbox"
      checked={when.allDay}
      onchange={(e) => (when = withAllDay(when, e.currentTarget.checked))}
    />
  </label>

  {#if when.allDay}
    <div class="grid grid-cols-2 gap-2">
      <label
        >From <input
          type="date"
          value={when.startDay}
          required
          onchange={(e) => e.currentTarget.value && (when = withStart(when, e.currentTarget.value))}
        /></label
      >
      <label
        >Until <input
          type="date"
          value={when.endDay}
          min={when.startDay}
          onchange={(e) => (when = { ...when, endDay: e.currentTarget.value || when.startDay })}
        /></label
      >
    </div>
  {:else}
    <div class="flex flex-col gap-2">
      <span class="text-sm font-medium">Starts</span>
      <div class="grid grid-cols-[1fr_auto] gap-2">
        <input
          type="date"
          aria-label="Start day"
          value={when.startDay}
          required
          onchange={(e) => e.currentTarget.value && (when = withStart(when, e.currentTarget.value))}
        />
        <input
          type="time"
          aria-label="Start time"
          class="w-36"
          value={when.startTime}
          required
          onchange={(e) =>
            e.currentTarget.value && (when = withStart(when, when.startDay, e.currentTarget.value))}
        />
      </div>
      <span class="text-sm font-medium"
        >Ends <span class="font-normal text-zinc-500">(optional)</span></span
      >
      <div class="grid grid-cols-[1fr_auto] gap-2">
        <input
          type="date"
          aria-label="End day"
          value={when.endDay}
          min={when.startDay}
          onchange={(e) => setEndDay(e.currentTarget.value)}
        />
        <input
          type="time"
          aria-label="End time"
          class="w-36"
          value={when.endTime}
          onchange={(e) => setEndTime(e.currentTarget.value)}
        />
      </div>
    </div>
  {/if}

  {#if wrong}
    <p class="text-sm text-red-600 dark:text-red-400" role="alert">{wrong}</p>
  {:else if length || moves}
    <div class="flex flex-wrap items-center gap-2">
      {#if length}<span class="text-xs font-medium text-zinc-500">{length}</span>{/if}
      {#if moves}
        <div class="ml-auto flex gap-1" role="group" aria-label="Move">
          {#each MOVES as m (m.days)}
            <button
              type="button"
              class="inline-flex items-center gap-1 rounded-full border border-zinc-200 bg-white px-2.5 py-1 text-xs font-semibold text-zinc-700 transition hover:border-zinc-300 active:scale-95 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200"
              aria-label="Move {m.label}"
              onclick={() => (when = shifted(when, m.days))}
              ><Icon name={m.days < 0 ? 'left' : 'right'} class="size-3" />{m.short}</button
            >
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  {@render children?.()}
</div>
