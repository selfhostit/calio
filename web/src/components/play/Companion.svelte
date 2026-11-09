<script lang="ts">
  import type { Species } from '../../lib/play/companion';
  import type { Mood } from '../../lib/play/mood';
  import RedPanda from './RedPanda.svelte';

  // Draws a companion of any kind; a new kind of animal is a new drawing here.
  let {
    species,
    mood,
    petted = false,
  }: { species: Species; mood: Mood; petted?: boolean } = $props();
</script>

{#if species === 'red_panda'}
  <RedPanda {mood} {petted} />
{/if}
