<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { ROLES, telLink, type Contact } from '../lib/contacts';
  import { mapLink } from '../lib/entries';
  import Icon from './Icon.svelte';

  let { onopen, onadd }: { onopen: (c: Contact) => void; onadd: () => void } = $props();

  const groups = $derived(
    ROLES.map((r) => ({
      ...r,
      contacts: calendar.contacts
        .filter((c) => c.role === r.role)
        .sort((a, b) => a.name.localeCompare(b.name)),
    })).filter((g) => g.contacts.length),
  );
</script>

{#if groups.length}
  <div class="flex flex-col gap-5">
    {#each groups as group (group.role)}
      <section class="flex flex-col gap-2">
        <h3 class="text-xs font-bold tracking-wider text-zinc-500 uppercase dark:text-zinc-400">
          {group.plural}
        </h3>
        {#each group.contacts as c (c.id)}
          <article class="card relative flex flex-col gap-2 p-4">
            <button
              class="absolute inset-0 rounded-2xl"
              onclick={() => onopen(c)}
              aria-label="Edit {c.name}"
            ></button>
            <div class="pointer-events-none flex items-center gap-3">
              <span
                class="grid size-10 shrink-0 place-items-center rounded-xl bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300"
                ><Icon name={group.icon} /></span
              >
              <div class="min-w-0 flex-1">
                <p class="truncate font-semibold">{c.name}</p>
                <p class="text-sm text-zinc-500">{group.label}</p>
              </div>
            </div>
            {#if c.phone || c.email || c.address}
              <div class="relative flex flex-wrap gap-2">
                {#if c.phone}
                  <a class="btn-secondary gap-1.5 px-3 py-1.5 text-sm" href={telLink(c.phone)}
                    ><Icon name="phone" class="size-4" />{c.phone}</a
                  >
                {/if}
                {#if c.email}
                  <a class="btn-secondary gap-1.5 px-3 py-1.5 text-sm" href="mailto:{c.email}"
                    ><Icon name="mail" class="size-4" />Email</a
                  >
                {/if}
                {#if c.address}
                  <a
                    class="btn-secondary gap-1.5 px-3 py-1.5 text-sm"
                    href={mapLink(c.address)}
                    target="_blank"
                    rel="noopener noreferrer"><Icon name="pin" class="size-4" />Map</a
                  >
                {/if}
              </div>
            {/if}
            {#if c.notes}
              <p
                class="pointer-events-none text-sm whitespace-pre-line text-zinc-600 dark:text-zinc-400"
              >
                {c.notes}
              </p>
            {/if}
          </article>
        {/each}
      </section>
    {/each}
  </div>
{:else if calendar.loaded}
  <div class="flex flex-col items-center gap-3 px-6 py-16 text-center">
    <span
      class="grid size-12 place-items-center rounded-2xl bg-zinc-100 text-zinc-500 dark:bg-zinc-800"
      ><Icon name="phone" /></span
    >
    <p class="mt-2 text-lg font-semibold">No contacts yet</p>
    <p class="text-sm text-zinc-500">
      Keep the vet, pet sitters and babysitters here, with their numbers and notes like feeding
      times or the door code.
    </p>
    <button class="btn-primary mt-2" onclick={onadd}>Add a contact</button>
  </div>
{/if}
