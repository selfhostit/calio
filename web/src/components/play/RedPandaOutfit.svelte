<svelte:options namespace="svg" />

<script lang="ts">
  import type { Slot } from '../../lib/play/shop';

  // What the red panda wears, drawn to fit its head and neck. An item from
  // the shop is drawn here by its id; only one item per slot is worn.
  let { wearing, slot }: { wearing: string[]; slot: Slot } = $props();

  const INK = '#2d1b16';
</script>

{#if slot === 'face' && wearing.includes('glasses')}
  <g fill="none" stroke={INK} stroke-width="2.5">
    <circle cx="81" cy="89" r="12" fill="#fff" fill-opacity="0.18" />
    <circle cx="119" cy="89" r="12" fill="#fff" fill-opacity="0.18" />
    <path d="M93 88 Q100 84 107 88 M69 87 L58 83 M131 87 L142 83" stroke-linecap="round" />
  </g>
{/if}

{#if slot === 'head'}
  {#if wearing.includes('party_hat')}
    <g transform="rotate(10 100 50)">
      <path d="M100 8 L120 54 L80 54 Z" fill="#f472b6" />
      <path d="M90 31 L110 31 L114 40 L86 40 Z M84 45 L116 45 L119 52 L81 52 Z" fill="#fde047" />
      <circle cx="100" cy="8" r="6" fill="#fde047" />
    </g>
  {:else if wearing.includes('flower_crown')}
    <path d="M58 60 Q100 38 142 60" fill="none" stroke="#65a30d" stroke-width="4" />
    {#each [[62, 57, '#f9a8d4'], [78, 49, '#fde047'], [100, 45, '#f9a8d4'], [122, 49, '#c4b5fd'], [138, 57, '#fde047']] as [x, y, c] (x)}
      <g transform="translate({x} {y})">
        {#each [0, 72, 144, 216, 288] as a (a)}
          <ellipse cx="0" cy="-5" rx="3.6" ry="5" fill={c as string} transform="rotate({a})" />
        {/each}
        <circle r="3" fill="#f59e0b" />
      </g>
    {/each}
  {:else if wearing.includes('beanie')}
    <path d="M50 66 Q52 28 100 26 Q148 28 150 66 Z" fill="#14b8a6" />
    <path
      d="M70 34 L72 64 M86 28 L87 64 M100 26 L100 64 M114 28 L113 64 M130 34 L128 64"
      stroke="#0f766e"
      stroke-width="2.5"
    />
    <rect x="46" y="58" width="108" height="14" rx="7" fill="#0d9488" />
    <circle cx="100" cy="22" r="9" fill="#fbefe2" />
  {:else if wearing.includes('explorer_hat')}
    <path d="M70 52 Q72 24 100 23 Q128 24 130 52 Z" fill="#d6b77a" />
    <rect x="70" y="42" width="60" height="9" fill="#7c5a2e" />
    <ellipse cx="100" cy="54" rx="50" ry="9" fill="#c9a764" />
  {/if}
{/if}

{#if slot === 'neck'}
  {#if wearing.includes('scarf')}
    <g>
      <path d="M62 132 Q100 148 138 132 L140 146 Q100 162 60 146 Z" fill="#ef4444" />
      <path
        d="M78 138 l4 15 M96 141 l2 15 M114 140 l-1 15 M130 136 l-3 14"
        stroke="#fef2f2"
        stroke-width="4"
      />
      <path d="M118 148 L126 176 L110 178 L106 152 Z" fill="#ef4444" />
      <path d="M113 160 l12 -2 M114 168 l12 -2" stroke="#fef2f2" stroke-width="3" />
    </g>
  {:else if wearing.includes('bandana')}
    <path d="M66 134 Q100 146 134 134 L100 168 Z" fill="#3b82f6" />
    {#each [[88, 144], [104, 147], [118, 142], [98, 157]] as [x, y] (`${x}${y}`)}
      <circle cx={x} cy={y} r="2.2" fill="#eff6ff" />
    {/each}
  {:else if wearing.includes('bow_tie')}
    <path d="M100 141 L80 131 L80 151 Z M100 141 L120 131 L120 151 Z" fill="#a855f7" />
    <rect x="94" y="135" width="12" height="12" rx="3" fill="#7e22ce" />
  {/if}
{/if}
