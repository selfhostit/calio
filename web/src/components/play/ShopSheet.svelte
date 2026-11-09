<script lang="ts">
  import { onMount } from 'svelte';
  import { errorMessage } from '../../lib/pb';
  import { play } from '../../lib/play/play.svelte';
  import type { Item } from '../../lib/play/shop';
  import Icon from '../Icon.svelte';
  import Sheet from '../Sheet.svelte';
  import Companion from './Companion.svelte';

  // Outfits for your companion, for berries. What you bought you can put on
  // and take off; one thing per spot (head, neck, face).
  let { onclose }: { onclose: () => void } = $props();

  // the balance may have changed on another device
  onMount(() => void play.refresh());

  let busy = $state('');
  let error = $state('');

  const berries = $derived(play.progress?.berries ?? 0);
  const ownedItem = (i: Item) => play.owned.find((o) => o.item === i.id);

  async function act(i: Item, fn: () => Promise<unknown>) {
    busy = i.id;
    error = '';
    try {
      await fn();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }
</script>

<Sheet title="Shop" {onclose}>
  <div class="flex flex-col gap-4 pt-1">
    <p class="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-400">
      <span
        class="inline-flex items-center gap-1 rounded-full bg-rose-100 px-2.5 py-1 font-semibold text-rose-700 dark:bg-rose-400/20 dark:text-rose-200"
        ><Icon name="berry" class="size-4" />{berries}</span
      >
      berries to spend. Every habit done earns one, a whole day three more.
    </p>
    {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}
    <ul class="grid grid-cols-2 gap-3">
      {#each play.items as i (i.id)}
        {@const owned = ownedItem(i)}
        <li class="card flex flex-col items-center gap-2 p-3 text-center">
          <div class="size-24">
            {#if play.companion}
              <Companion species={play.companion.species} mood="happy" wearing={[i.id]} />
            {/if}
          </div>
          <p class="text-sm font-semibold">{i.name}</p>
          {#if owned}
            <button
              class="w-full rounded-xl px-3 py-2 text-sm font-semibold transition {owned.equipped
                ? 'bg-amber-500 text-white'
                : 'bg-zinc-100 dark:bg-zinc-800'}"
              disabled={busy === i.id}
              aria-pressed={owned.equipped}
              onclick={() => act(i, () => play.wear(i, !owned.equipped))}
              >{owned.equipped ? 'Wearing' : 'Put on'}</button
            >
          {:else}
            <button
              class="inline-flex w-full items-center justify-center gap-1 rounded-xl border px-3 py-2 text-sm font-semibold transition disabled:opacity-50 {berries >=
              i.price
                ? 'border-rose-300 text-rose-700 hover:bg-rose-50 dark:border-rose-400/40 dark:text-rose-200 dark:hover:bg-rose-400/10'
                : 'border-zinc-200 text-zinc-500 dark:border-zinc-700'}"
              disabled={busy === i.id || berries < i.price}
              onclick={() => act(i, () => play.buy(i))}
              ><Icon name="berry" class="size-4" />{i.price}</button
            >
          {/if}
        </li>
      {/each}
    </ul>
  </div>
</Sheet>
