<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { install } from '../lib/install.svelte';
  import { pb } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Avatar from './Avatar.svelte';
  import Categories from './Categories.svelte';
  import Feed from './Feed.svelte';
  import Imports from './Imports.svelte';
  import Notifications from './Notifications.svelte';
  import Sheet from './Sheet.svelte';

  let { onclose, importing = false }: { onclose: () => void; importing?: boolean } = $props();

  // opened to import calendars: go straight there
  $effect(() => {
    if (importing) document.getElementById('import')?.scrollIntoView({ block: 'start' });
  });

  const user = $derived(session.user!);
  const version = pb
    .send<{ version: string }>('/api/app/info', {})
    .then((i) => i.version)
    .catch(() => '');
</script>

<Sheet title="Settings" {onclose}>
  <div class="flex flex-col gap-4 pt-1">
    <section class="card flex items-center gap-3">
      <Avatar person={user} size="size-11 text-base" />
      <div class="min-w-0 flex-1">
        <p class="truncate font-semibold">{user.name || user.email}</p>
        <p class="truncate text-sm text-zinc-500">
          {user.email}{session.isOwner ? ' · owner' : ''}
        </p>
      </div>
      <button class="btn-secondary px-3 py-2 text-sm" onclick={() => session.logout()}
        >Sign out</button
      >
    </section>

    {#if install.canPrompt}
      <section class="card flex items-center justify-between gap-4">
        <p>Install Calio for quick access from your home screen.</p>
        <button class="btn-primary shrink-0" onclick={() => install.install()}>Install</button>
      </section>
    {:else if install.needsIosSteps}
      <section class="card">
        <p>To install Calio, tap <b>Share</b> in Safari, then <b>Add to Home Screen</b>.</p>
      </section>
    {/if}

    <Categories />

    <Notifications />

    <Imports />

    <Feed />

    <section class="card flex flex-col gap-2 text-sm">
      <h3 class="font-semibold">Household</h3>
      <div class="flex flex-wrap gap-2">
        {#each calendar.people as p (p.id)}
          <span
            class="inline-flex items-center gap-2 rounded-full bg-zinc-100 py-1 pr-3 pl-1 dark:bg-zinc-800"
          >
            <Avatar person={p} size="size-6 text-[10px]" />{p.name || p.email || 'No name yet'}
          </span>
        {/each}
      </div>
      {#if session.isOwner}
        <p class="text-zinc-500">
          Add people in the <a class="text-brand-600 underline dark:text-brand-400" href="/_/"
            >admin page</a
          >, under users.
        </p>
      {/if}
    </section>

    {#await version then v}
      {#if v}<p class="text-center text-xs text-zinc-400">Calio {v}</p>{/if}
    {/await}
  </div>
</Sheet>
