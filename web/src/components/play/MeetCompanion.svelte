<script lang="ts">
  import { errorMessage } from '../../lib/pb';
  import { SPECIES } from '../../lib/play/companion';
  import { STARTERS } from '../../lib/play/habits';
  import { play } from '../../lib/play/play.svelte';
  import Icon from '../Icon.svelte';
  import Companion from './Companion.svelte';

  // The first visit: meet your red panda, name it and pick a few habits.
  const species = 'red_panda';
  let name = $state(SPECIES[species].name);
  let picked = $state(STARTERS.slice(0, 3).map((s) => s.title));
  let busy = $state(false);
  let error = $state('');

  const toggle = (t: string) =>
    (picked = picked.includes(t) ? picked.filter((x) => x !== t) : [...picked, t]);

  async function meet(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    try {
      await play.meet(
        name,
        species,
        STARTERS.filter((s) => picked.includes(s.title)),
      );
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form class="card flex flex-col gap-5 p-5" onsubmit={meet}>
  <div class="flex flex-col items-center gap-2 text-center">
    <div class="size-40"><Companion {species} mood="happy" /></div>
    <h2 class="text-2xl font-bold tracking-tight">Meet your red panda</h2>
    <p class="max-w-sm text-sm text-zinc-600 dark:text-zinc-400">
      Tick off a few small habits each day and it stays happy. Skip them for a while and it gets a
      bit droopy, but one good day cheers it right up.
    </p>
  </div>

  <label>
    Its name
    <input bind:value={name} required maxlength="30" />
  </label>

  <fieldset class="flex flex-col gap-2">
    <legend class="mb-2 text-sm font-medium">Start with</legend>
    <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
      {#each STARTERS as s (s.title)}
        <button
          type="button"
          aria-pressed={picked.includes(s.title)}
          class="flex items-center gap-3 rounded-xl border px-3 py-2.5 text-left text-sm font-medium transition {picked.includes(
            s.title,
          )
            ? 'border-amber-400 bg-amber-50 dark:border-amber-500/60 dark:bg-amber-400/10'
            : 'border-zinc-200 dark:border-zinc-700'}"
          onclick={() => toggle(s.title)}
        >
          <Icon name={s.icon} class="size-5 text-amber-700 dark:text-amber-300" />
          <span class="flex-1">{s.title}</span>
          {#if picked.includes(s.title)}<Icon name="check" class="size-4 text-amber-600" />{/if}
        </button>
      {/each}
    </div>
  </fieldset>

  {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}
  <button class="btn-primary" disabled={busy}>Say hi to {name.trim() || 'your panda'}</button>
</form>
