<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { fileUrl, isImage, type Doc } from '../lib/documents';
  import Icon from './Icon.svelte';

  let { onopen, onadd }: { onopen: (d: Doc) => void; onadd: () => void } = $props();

  const docs = $derived(
    [...calendar.documents].sort((a, b) =>
      a.title.localeCompare(b.title, undefined, { sensitivity: 'base' }),
    ),
  );

  const preview = (d: Doc) =>
    d.body
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
      .slice(0, 2)
      .join(' · ');
</script>

{#if docs.length}
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
    {#each docs as d (d.id)}
      {@const photo = d.files?.find(isImage)}
      <button
        class="card flex items-start gap-3 p-4 text-left transition hover:bg-zinc-50 active:scale-[0.99] dark:hover:bg-zinc-800/60"
        onclick={() => onopen(d)}
      >
        {#if photo}
          <img
            src={fileUrl(d, photo, '400x400')}
            alt=""
            class="size-12 shrink-0 rounded-xl object-cover"
            loading="lazy"
          />
        {:else}
          <span
            class="grid size-12 shrink-0 place-items-center rounded-xl bg-brand-500/10 text-brand-700 dark:text-brand-400"
            ><Icon name={d.icon || 'doc'} /></span
          >
        {/if}
        <span class="flex min-w-0 flex-1 flex-col gap-0.5">
          <span class="flex items-center gap-2">
            <span class="truncate font-semibold">{d.title}</span>
            {#if d.shared}
              <span
                class="inline-flex shrink-0 items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-800 dark:bg-emerald-400/10 dark:text-emerald-300"
                ><Icon name="share" class="size-3" />Shared</span
              >
            {/if}
          </span>
          {#if preview(d)}
            <span class="line-clamp-2 text-sm text-zinc-500 dark:text-zinc-400">{preview(d)}</span>
          {/if}
          {#if d.files?.length}
            <span class="inline-flex items-center gap-1 text-xs text-zinc-400"
              ><Icon name="clip" class="size-3" />{d.files.length}
              {d.files.length === 1 ? 'file' : 'files'}</span
            >
          {/if}
        </span>
      </button>
    {/each}
  </div>
{:else if calendar.loaded}
  <div class="flex flex-col items-center gap-3 px-6 py-16 text-center">
    <div class="flex gap-2">
      <span
        class="grid size-12 -rotate-6 place-items-center rounded-2xl bg-amber-50 text-amber-700 dark:bg-amber-400/10 dark:text-amber-300"
        ><Icon name="paw" /></span
      >
      <span
        class="grid size-12 place-items-center rounded-2xl bg-brand-500/10 text-brand-700 dark:text-brand-400"
        ><Icon name="doc" /></span
      >
      <span
        class="grid size-12 rotate-6 place-items-center rounded-2xl bg-sky-50 text-sky-600 dark:bg-sky-400/10 dark:text-sky-300"
        ><Icon name="home" /></span
      >
    </div>
    <p class="mt-2 text-lg font-semibold">No documents yet</p>
    <p class="max-w-sm text-sm text-zinc-500">
      Write down what a sitter needs to know: a page about the dog, the house or the kids, with
      photos and the vet's number. Share it with a private link.
    </p>
    <button class="btn-primary mt-2" onclick={onadd}>Write a document</button>
  </div>
{/if}
