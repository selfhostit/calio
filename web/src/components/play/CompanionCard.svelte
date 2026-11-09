<script lang="ts">
  import type { Companion as Pal } from '../../lib/play/companion';
  import { MOODS, type Mood } from '../../lib/play/mood';
  import Icon from '../Icon.svelte';
  import Companion from './Companion.svelte';

  // Your companion: tap it to pet it. It says one thing a day, and shows how
  // much of today is done.
  let {
    companion,
    mood,
    line,
    done,
    due,
  }: { companion: Pal; mood: Mood; line: string; done: number; due: number } = $props();

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
      <Companion species={companion.species} {mood} {petted} />
      {#each hearts as h (h)}
        <span class="heart pointer-events-none absolute top-6 left-1/2 text-rose-500"
          ><Icon name="heart" class="size-6 fill-current" /></span
        >
      {/each}
    </button>
    <div class="flex min-w-0 flex-1 flex-col gap-2">
      <div>
        <h2 class="truncate text-xl font-bold tracking-tight">{companion.name}</h2>
        <p class="text-sm font-medium text-amber-700 dark:text-amber-300">{MOODS[mood].label}</p>
      </div>
      <p
        class="bubble relative rounded-2xl bg-white px-3 py-2 text-sm shadow-sm ring-1 ring-zinc-200 dark:bg-zinc-800 dark:ring-zinc-700"
      >
        {line}
      </p>
    </div>
  </div>
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
