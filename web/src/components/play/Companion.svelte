<script lang="ts">
  import type { Species } from '../../lib/play/companion';
  import type { MomentKind } from '../../lib/play/moments';
  import type { Mood } from '../../lib/play/mood';
  import type { Stage } from '../../lib/play/shop';
  import RedPanda from './RedPanda.svelte';

  // Draws a companion of any kind; a new kind of animal is a new drawing here.
  let {
    species,
    mood,
    petted = false,
    wearing = [],
    stage = 'grown',
    moment = null,
  }: {
    species: Species;
    mood: Mood;
    petted?: boolean;
    wearing?: string[];
    stage?: Stage;
    moment?: MomentKind | null;
  } = $props();
</script>

{#if species === 'red_panda'}
  <RedPanda {mood} {petted} {wearing} {stage} {moment} />
{/if}
