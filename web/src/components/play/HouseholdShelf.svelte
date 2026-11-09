<script lang="ts">
  import { onMount } from 'svelte';
  import { calendar, firstName } from '../../lib/calendar.svelte';
  import { errorMessage } from '../../lib/pb';
  import { household } from '../../lib/play/household.svelte';
  import { MOODS } from '../../lib/play/mood';
  import Icon from '../Icon.svelte';
  import Companion from './Companion.svelte';

  // Everyone else's companion: how their day is going, and a cheer to send.
  onMount(() => {
    household.start();
  });

  let error = $state('');

  async function cheer(user: string) {
    error = '';
    try {
      await household.cheer(user);
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if household.pals.length}
  <section class="flex flex-col gap-2">
    <h3 class="px-1 text-xs font-semibold tracking-wide text-zinc-500 uppercase">The household</h3>
    {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}
    <ul class="-mx-4 flex gap-3 overflow-x-auto px-4 pb-1">
      {#each household.pals as p (p.companion.id)}
        <li class="card flex w-40 shrink-0 flex-col items-center gap-1.5 p-3 text-center">
          <div class="size-24">
            <Companion
              species={p.companion.species}
              mood={p.mood}
              wearing={p.wearing}
              stage={p.stage}
            />
          </div>
          <p class="w-full truncate font-semibold">{p.companion.name}</p>
          <p class="w-full truncate text-xs text-zinc-500 dark:text-zinc-400">
            {firstName(calendar.person(p.companion.user))} · {MOODS[p.mood].label.toLowerCase()}
          </p>
          {#if p.due}
            <p class="text-xs font-medium text-amber-700 dark:text-amber-300">
              {p.done} of {p.due} today
            </p>
          {/if}
          <button
            class="mt-1 inline-flex w-full items-center justify-center gap-1.5 rounded-xl px-3 py-2 text-sm font-semibold transition {p.cheered
              ? 'bg-rose-100 text-rose-700 dark:bg-rose-400/20 dark:text-rose-200'
              : 'border border-zinc-300 hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-800'}"
            disabled={p.cheered}
            onclick={() => cheer(p.companion.user)}
            ><Icon name="heart" class="size-4 {p.cheered ? 'fill-current' : ''}" />{p.cheered
              ? 'Cheered'
              : 'Cheer on'}</button
          >
        </li>
      {/each}
    </ul>
  </section>
{/if}
