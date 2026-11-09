<script lang="ts">
  import type { Companion as Pal } from '../../lib/play/companion';
  import { MOODS, type Mood } from '../../lib/play/mood';
  import { STAGES, type Progress } from '../../lib/play/shop';
  import Icon from '../Icon.svelte';
  import Companion from './Companion.svelte';

  // Your companion: tap it to pet it. It says one thing a day, and shows how
  // much of today is done, your berries and streak, and how it's growing.
  let {
    companion,
    mood,
    line,
    done,
    due,
    progress,
    wearing,
    onshop,
  }: {
    companion: Pal;
    mood: Mood;
    line: string;
    done: number;
    due: number;
    progress: Progress | null;
    wearing: string[];
    onshop: () => void;
  } = $props();

  // a "+3" floats up when berries come in
  let gained = $state<{ id: number; n: number } | null>(null);
  let last: number | null = null;
  $effect(() => {
    const now = progress?.berries;
    if (now === undefined) return;
    if (last !== null && now > last) gained = { id: Date.now(), n: now - last };
    last = now;
  });

  let petted = $state(false);
  let hearts = $state<number[]>([]);
  let timer: ReturnType<typeof setTimeout> | undefined;

  function pet() {
    petted = true;
    const id = Date.now();
    hearts = [...hearts.slice(-4), id];
    clearTimeout(timer);
    timer = setTimeout(() => (petted = false), 1100);
    setTimeout(() => (hearts = hearts.filter((h) => h !== id)), 1200);
  }

  const share = $derived(due ? done / due : 0);
</script>

<section
  class="card relative overflow-hidden bg-gradient-to-b from-amber-50 to-white p-5 dark:from-amber-950/30 dark:to-zinc-900"
>
  <div class="flex items-center gap-4">
    <button
      class="relative size-36 shrink-0 touch-manipulation sm:size-44"
      onclick={pet}
      aria-label="Pet {companion.name}"
    >
      <Companion
        species={companion.species}
        {mood}
        {petted}
        {wearing}
        stage={progress?.stage ?? 'grown'}
      />
      {#each hearts as h (h)}
        <span class="heart pointer-events-none absolute top-6 left-1/2 text-rose-500"
          ><Icon name="heart" class="size-6 fill-current" /></span
        >
      {/each}
    </button>
    <div class="flex min-w-0 flex-1 flex-col gap-2">
      <div>
        <h2 class="truncate text-xl font-bold tracking-tight">{companion.name}</h2>
        <p class="text-sm font-medium text-amber-700 dark:text-amber-300">
          {MOODS[mood].label}
          {#if progress && progress.stage !== 'grown'}
            <span class="text-zinc-500 dark:text-zinc-400">&middot; {STAGES[progress.stage]}</span>
          {/if}
        </p>
      </div>
      <p
        class="bubble relative rounded-2xl bg-white px-3 py-2 text-sm shadow-sm ring-1 ring-zinc-200 dark:bg-zinc-800 dark:ring-zinc-700"
      >
        {line}
      </p>
    </div>
  </div>
  {#if progress}
    <div class="mt-4 flex flex-wrap items-center gap-2 text-sm font-semibold">
      <span
        class="relative inline-flex items-center gap-1 rounded-full bg-rose-100 px-2.5 py-1 text-rose-700 dark:bg-rose-400/20 dark:text-rose-200"
        title="Berries"
      >
        <Icon name="berry" class="size-4" />{progress.berries}<span class="sr-only"> berries</span>
        {#key gained?.id}
          {#if gained}
            <span class="gain pointer-events-none absolute -top-1 right-0 text-rose-600"
              >+{gained.n}</span
            >
          {/if}
        {/key}
      </span>
      {#if progress.streak > 1}
        <span
          class="inline-flex items-center gap-1 rounded-full bg-orange-100 px-2.5 py-1 text-orange-700 dark:bg-orange-400/20 dark:text-orange-200"
          ><Icon name="flame" class="size-4" />{progress.streak} days in a row</span
        >
      {/if}
      <button
        class="ml-auto inline-flex items-center gap-1.5 rounded-full border border-zinc-300 px-3 py-1 hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-800"
        onclick={onshop}><Icon name="bag" class="size-4" />Shop</button
      >
    </div>
    {#if progress.next_stage}
      <p class="mt-2 text-xs text-zinc-500 dark:text-zinc-400">
        {progress.next_stage - progress.lifetime} more habits until {companion.name} grows.
      </p>
    {/if}
  {/if}
  {#if due}
    <div class="mt-4 flex items-center gap-3 text-sm">
      <div
        class="h-2.5 flex-1 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-800"
        role="progressbar"
        aria-valuemin="0"
        aria-valuemax={due}
        aria-valuenow={done}
        aria-label="Done today"
      >
        <div
          class="h-full rounded-full bg-amber-500 transition-[width] duration-500"
          style="width: {share * 100}%"
        ></div>
      </div>
      <span class="shrink-0 font-medium text-zinc-600 dark:text-zinc-300"
        >{done} of {due} today</span
      >
    </div>
  {/if}
</section>

<style>
  .bubble::before {
    content: '';
    position: absolute;
    left: -6px;
    top: 50%;
    width: 12px;
    height: 12px;
    background: inherit;
    transform: translateY(-50%) rotate(45deg);
    border-left: 1px solid rgb(0 0 0 / 0.06);
    border-bottom: 1px solid rgb(0 0 0 / 0.06);
  }
  .heart {
    animation: float 1.2s ease-out forwards;
  }
  .gain {
    animation: gain 1.2s ease-out forwards;
  }
  @keyframes gain {
    from {
      transform: translateY(0);
      opacity: 1;
    }
    to {
      transform: translateY(-24px);
      opacity: 0;
    }
  }
  @keyframes float {
    from {
      transform: translate(-50%, 0) scale(0.6);
      opacity: 1;
    }
    to {
      transform: translate(calc(-50% + 14px), -70px) scale(1.1);
      opacity: 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .heart {
      animation-duration: 0.01s;
    }
  }
</style>
