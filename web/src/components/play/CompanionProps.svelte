<svelte:options namespace="svg" />

<script lang="ts">
  import type { MomentKind } from '../../lib/play/moments';

  // What sits beside a companion for a calendar moment: a suitcase before
  // and during a trip, a visiting puppy on pet-sitting days, confetti on a
  // birthday. Drawn in the companion's 200 by 200 space, to its left.
  let { moment }: { moment: MomentKind } = $props();
</script>

{#if moment === 'packing' || moment === 'away'}
  <g class="prop">
    <path
      d="M28 150 v-8 a4 4 0 0 1 4 -4 h12 a4 4 0 0 1 4 4 v8"
      fill="none"
      stroke="#7c5a2e"
      stroke-width="4"
    />
    <rect x="14" y="150" width="48" height="38" rx="7" fill="#f59e0b" />
    <rect x="14" y="166" width="48" height="5" fill="#b45309" />
    <circle cx="26" cy="159" r="4.5" fill="#38bdf8" />
    <rect x="42" y="175" width="13" height="8" rx="2" fill="#f472b6" />
    {#if moment === 'away'}
      <path d="M40 120 l18 -6 -4 -6 8 -2 6 6 10 -3 a3 3 0 0 1 2 5 l-28 12 Z" fill="#38bdf8" />
    {/if}
  </g>
{:else if moment === 'visitor'}
  <g class="prop">
    <ellipse cx="38" cy="176" rx="20" ry="13" fill="#e7c9a0" />
    <circle cx="30" cy="154" r="14" fill="#e7c9a0" />
    <ellipse cx="18" cy="152" rx="5" ry="10" fill="#a0784c" transform="rotate(20 18 152)" />
    <ellipse cx="42" cy="152" rx="5" ry="10" fill="#a0784c" transform="rotate(-20 42 152)" />
    <circle cx="25" cy="153" r="2.4" fill="#2d1b16" />
    <circle cx="35" cy="153" r="2.4" fill="#2d1b16" />
    <ellipse cx="30" cy="160" rx="3.4" ry="2.4" fill="#2d1b16" />
    <path
      d="M56 172 q8 -8 6 -16"
      fill="none"
      stroke="#e7c9a0"
      stroke-width="5"
      stroke-linecap="round"
      class="wag"
    />
  </g>
{:else if moment === 'birthday'}
  <g class="confetti">
    {#each [[20, 40, '#f472b6', 20], [176, 60, '#fde047', -30], [24, 110, '#38bdf8', 45], [172, 20, '#a78bfa', 10], [40, 18, '#34d399', -15], [182, 130, '#f472b6', 60]] as [x, y, c, r] (`${x}${y}`)}
      <rect
        x={x as number}
        y={y as number}
        width="8"
        height="4"
        rx="1"
        fill={c as string}
        transform="rotate({r} {x} {y})"
      />
    {/each}
  </g>
{/if}

<style>
  .prop {
    animation: pop 0.5s ease-out;
    transform-origin: 38px 188px;
    transform-box: view-box;
  }
  .wag {
    transform-origin: 56px 172px;
    transform-box: view-box;
    animation: wag 0.6s ease-in-out infinite;
  }
  .confetti {
    animation: fall 3s ease-in-out infinite;
  }
  @keyframes pop {
    from {
      transform: scale(0.6);
      opacity: 0;
    }
  }
  @keyframes wag {
    50% {
      transform: rotate(25deg);
    }
  }
  @keyframes fall {
    50% {
      transform: translateY(6px);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .prop,
    .wag,
    .confetti {
      animation: none;
    }
  }
</style>
