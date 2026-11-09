<script lang="ts">
  import type { Mood } from '../../lib/play/mood';
  import type { Stage } from '../../lib/play/shop';
  import type { MomentKind } from '../../lib/play/moments';
  import CompanionProps from './CompanionProps.svelte';
  import RedPandaOutfit from './RedPandaOutfit.svelte';

  // The red panda, drawn in parts (ears, head, face, body, tail) so a mood is
  // a different face and pose, not a different drawing. `petted` shows it
  // being petted: eyes closed, a happy wiggle. It wears what's in `wearing`
  // (RedPandaOutfit.svelte) and is smaller while it's young.
  let {
    mood,
    petted = false,
    wearing = [],
    stage = 'grown',
    moment = null,
  }: {
    mood: Mood;
    petted?: boolean;
    wearing?: string[];
    stage?: Stage;
    moment?: MomentKind | null;
  } = $props();

  // a birthday is a party: the party hat goes on, whatever hat was on
  const HATS = ['party_hat', 'flower_crown', 'beanie', 'explorer_hat'];
  const worn = $derived(
    moment === 'birthday' ? ['party_hat', ...wearing.filter((w) => !HATS.includes(w))] : wearing,
  );

  const size = $derived({ cub: 0.8, young: 0.9, grown: 1 }[stage]);

  const face = $derived(petted ? 'glowing' : mood);
  // ears and tail droop as the mood drops
  const ears = $derived({ glowing: -6, happy: 0, okay: 4, droopy: 16, sad: 26 }[face]);
  const blush = $derived(face === 'glowing' || face === 'happy');

  const FUR = '#cf5f2c';
  const FUR_DARK = '#9b3c1b';
  const CREAM = '#fbefe2';
  const INK = '#2d1b16';
  const LEGS = '#4a2a20';
</script>

<svg
  viewBox="0 0 200 200"
  class="panda mood-{mood}"
  class:petted
  role="img"
  aria-label="A red panda, looking {face === 'glowing' ? 'delighted' : face}"
>
  <g style="transform: scale({size}); transform-origin: 100px 196px">
    <!-- tail: a thick ringed stroke curling up behind the body -->
    <g class="tail">
      <path
        d="M120 158 C160 164 188 136 176 94"
        fill="none"
        stroke={FUR}
        stroke-width="26"
        stroke-linecap="round"
      />
      <path
        d="M120 158 C160 164 188 136 176 94"
        fill="none"
        stroke={FUR_DARK}
        stroke-width="26"
        stroke-dasharray="9 11"
        stroke-dashoffset="-6"
      />
    </g>

    <g class="body">
      <ellipse cx="100" cy="152" rx="42" ry="33" fill={FUR} />
      <ellipse cx="100" cy="166" rx="30" ry="18" fill={LEGS} />
      <ellipse cx="80" cy="183" rx="14" ry="9" fill={LEGS} />
      <ellipse cx="120" cy="183" rx="14" ry="9" fill={LEGS} />
      <ellipse cx="68" cy="152" rx="10" ry="15" fill={LEGS} transform="rotate(18 68 152)" />
      <ellipse cx="132" cy="152" rx="10" ry="15" fill={LEGS} transform="rotate(-18 132 152)" />
    </g>

    <g class="head">
      <g style="transform: rotate({-ears}deg); transform-origin: 66px 70px" class="ear">
        <ellipse cx="58" cy="50" rx="17" ry="20" fill={CREAM} transform="rotate(-22 58 50)" />
        <ellipse cx="59" cy="52" rx="10" ry="13" fill={LEGS} transform="rotate(-22 59 52)" />
      </g>
      <g style="transform: rotate({ears}deg); transform-origin: 134px 70px" class="ear">
        <ellipse cx="142" cy="50" rx="17" ry="20" fill={CREAM} transform="rotate(22 142 50)" />
        <ellipse cx="141" cy="52" rx="10" ry="13" fill={LEGS} transform="rotate(22 141 52)" />
      </g>

      <ellipse cx="100" cy="92" rx="57" ry="46" fill={FUR} />
      <!-- white cheeks, muzzle and brows -->
      <ellipse cx="60" cy="106" rx="16" ry="12" fill={CREAM} />
      <ellipse cx="140" cy="106" rx="16" ry="12" fill={CREAM} />
      <ellipse cx="100" cy="112" rx="25" ry="17" fill={CREAM} />
      <ellipse
        cx="80"
        cy="73"
        rx="9"
        ry="5"
        fill={CREAM}
        transform="rotate({face === 'sad' || face === 'droopy' ? 14 : -6} 80 73)"
      />
      <ellipse
        cx="120"
        cy="73"
        rx="9"
        ry="5"
        fill={CREAM}
        transform="rotate({face === 'sad' || face === 'droopy' ? -14 : 6} 120 73)"
      />
      <!-- the russet marks that run down from the eyes -->
      <path
        d="M80 96 Q77 106 72 114 M120 96 Q123 106 128 114"
        stroke={FUR_DARK}
        stroke-width="5"
        stroke-linecap="round"
        fill="none"
      />

      {#if blush}
        <ellipse cx="68" cy="113" rx="7" ry="4" fill="#f28b9b" opacity="0.55" />
        <ellipse cx="132" cy="113" rx="7" ry="4" fill="#f28b9b" opacity="0.55" />
      {/if}

      <!-- eyes -->
      {#if face === 'glowing'}
        <path
          d="M73 91 Q81 82 89 91 M111 91 Q119 82 127 91"
          stroke={INK}
          stroke-width="4"
          stroke-linecap="round"
          fill="none"
        />
      {:else if face === 'droopy'}
        <ellipse cx="81" cy="92" rx="7" ry="4.5" fill={INK} />
        <ellipse cx="119" cy="92" rx="7" ry="4.5" fill={INK} />
        <path
          d="M72 89 L90 89 M110 89 L128 89"
          stroke={FUR_DARK}
          stroke-width="4"
          stroke-linecap="round"
        />
      {:else}
        <g class="eyes">
          <circle cx="81" cy={face === 'sad' ? 92 : 89} r="7.5" fill={INK} />
          <circle cx="119" cy={face === 'sad' ? 92 : 89} r="7.5" fill={INK} />
          <circle cx={face === 'sad' ? 79 : 84} cy={face === 'sad' ? 95 : 86} r="2.6" fill="#fff" />
          <circle
            cx={face === 'sad' ? 117 : 122}
            cy={face === 'sad' ? 95 : 86}
            r="2.6"
            fill="#fff"
          />
        </g>
      {/if}
      {#if face === 'sad'}
        <path class="tear" d="M74 101 q-3 5 0 7 q3 -2 0 -7Z" fill="#7cc4f4" />
      {/if}

      <!-- nose and mouth -->
      <ellipse cx="100" cy="104" rx="6.5" ry="4.5" fill={INK} />
      {#if face === 'glowing'}
        <path d="M89 112 Q100 128 111 112 Z" fill={INK} />
        <path d="M94 118 Q100 124 106 118 Q100 116 94 118Z" fill="#f28b9b" />
      {:else}
        <path
          d={{
            happy: 'M91 112 Q100 121 109 112',
            okay: 'M94 114 Q100 117 106 114',
            droopy: 'M94 116 L106 116',
            sad: 'M92 119 Q100 111 108 119',
          }[face]}
          stroke={INK}
          stroke-width="3"
          stroke-linecap="round"
          fill="none"
        />
      {/if}
      <RedPandaOutfit wearing={worn} slot="face" />
      <RedPandaOutfit wearing={worn} slot="head" />
    </g>

    <g class="body">
      <RedPandaOutfit wearing={worn} slot="neck" />
    </g>
  </g>

  {#if moment}
    <CompanionProps {moment} />
  {/if}

  {#if mood === 'glowing'}
    <g class="sparkles" fill="#fbbf24">
      <path d="M28 52 l3 8 8 3 -8 3 -3 8 -3 -8 -8 -3 8 -3Z" />
      <path d="M172 34 l2 6 6 2 -6 2 -2 6 -2 -6 -6 -2 6 -2Z" />
      <path d="M30 140 l2 5 5 2 -5 2 -2 5 -2 -5 -5 -2 5 -2Z" />
    </g>
  {/if}
</svg>

<style>
  .panda {
    overflow: visible;
  }
  .panda g {
    transform-box: view-box;
  }
  .tail {
    transform-origin: 120px 158px;
    animation: sway 3.2s ease-in-out infinite;
  }
  .body,
  .head {
    transform-origin: 100px 186px;
    animation: breathe 3.6s ease-in-out infinite;
  }
  .ear {
    transition: transform 0.6s ease;
  }
  .eyes {
    transform-origin: 100px 89px;
    animation: blink 5s infinite;
  }
  .mood-glowing .head,
  .mood-glowing .body {
    animation: bounce 1.4s ease-in-out infinite;
  }
  .mood-droopy .tail,
  .mood-sad .tail {
    animation-duration: 6s;
  }
  .mood-sad .head {
    transform: translateY(3px) rotate(-3deg);
    animation: none;
  }
  .tear {
    animation: tear 2.4s ease-in infinite;
  }
  .sparkles {
    transform-origin: 100px 100px;
    animation: twinkle 1.8s ease-in-out infinite;
  }
  .petted .head,
  .petted .body {
    animation: wiggle 0.5s ease-in-out 2;
  }
  .petted .tail {
    animation: sway 0.4s ease-in-out 3;
  }

  @keyframes sway {
    50% {
      transform: rotate(-7deg);
    }
  }
  @keyframes breathe {
    50% {
      transform: scale(1.015, 0.985);
    }
  }
  @keyframes bounce {
    50% {
      transform: translateY(-5px);
    }
  }
  @keyframes wiggle {
    25% {
      transform: rotate(-5deg);
    }
    75% {
      transform: rotate(5deg);
    }
  }
  @keyframes blink {
    0%,
    95%,
    100% {
      transform: scaleY(1);
    }
    97% {
      transform: scaleY(0.1);
    }
  }
  @keyframes tear {
    0% {
      transform: translateY(0);
      opacity: 0;
    }
    20% {
      opacity: 1;
    }
    100% {
      transform: translateY(14px);
      opacity: 0;
    }
  }
  @keyframes twinkle {
    50% {
      opacity: 0.4;
      transform: scale(0.96);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .panda * {
      animation: none !important;
    }
  }
</style>
