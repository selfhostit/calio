<script lang="ts">
  import { untrack } from 'svelte';
  import { calendar } from '../lib/calendar.svelte';
  import { TEMPLATES, fileLabel, fileUrl, isImage, linkParts, type Doc } from '../lib/documents';
  import { ICONS } from '../lib/look';
  import { errorMessage, pb } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Icon from './Icon.svelte';
  import Sheet from './Sheet.svelte';

  // A document opens to read (with its photos, links and share link); Edit
  // changes it. A new one opens to write.
  let { doc = null, onclose }: { doc?: Doc | null; onclose: () => void } = $props();

  const opened = untrack(() => doc);
  // the live copy, so a change from the other phone shows here too
  const current = $derived(
    opened ? (calendar.documents.find((d) => d.id === opened.id) ?? opened) : null,
  );

  let editing = $state(!opened);
  let title = $state('');
  let icon = $state('doc');
  let body = $state('');
  let added = $state<File[]>([]);
  let removed = $state<string[]>([]);
  let busy = $state(false);
  let error = $state('');

  function edit() {
    title = current?.title ?? '';
    icon = current?.icon || 'doc';
    body = current?.body ?? '';
    added = [];
    removed = [];
    error = '';
    editing = true;
  }
  if (!opened) edit();

  const canDelete = $derived(
    !!current && (current.created_by === session.user?.id || session.isOwner),
  );
  const keptFiles = $derived((current?.files ?? []).filter((f) => !removed.includes(f)));

  function useTemplate(t: (typeof TEMPLATES)[number]) {
    if (!title.trim() || TEMPLATES.some((x) => x.title === title)) title = t.title;
    icon = t.icon;
    if (!body.trim() || TEMPLATES.some((x) => x.body === body)) body = t.body;
  }

  function pickFiles(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    added = [...added, ...Array.from(input.files ?? [])];
    input.value = '';
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    const data = new FormData();
    data.append('title', title.trim());
    data.append('icon', icon);
    data.append('body', body.trim());
    for (const f of added) data.append('files+', f);
    for (const name of removed) data.append('files-', name);
    try {
      await calendar.saveDocument(current?.id ?? null, data);
      if (current) editing = false;
      else onclose();
    } catch (err) {
      const fields = (err as { response?: { data?: Record<string, { message: string }> } }).response
        ?.data;
      error = (fields && Object.values(fields)[0]?.message) || errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove() {
    if (!current || !confirm(`Delete “${current.title}”?`)) return;
    busy = true;
    try {
      await calendar.removeDocument(current.id);
      onclose();
    } catch (err) {
      error = errorMessage(err);
      busy = false;
    }
  }

  // ---- The private link ----

  let link = $state<{ url: string; public: boolean } | null>(null);
  let shareMessage = $state('');

  $effect(() => {
    if (current?.shared && !link) getLink('GET');
  });

  async function getLink(method: 'GET' | 'POST' | 'DELETE') {
    if (!current) return;
    if (
      method === 'DELETE' &&
      !confirm('Stop sharing? The link stops working for anyone who has it.')
    )
      return;
    shareMessage = '';
    try {
      link = await pb.send(`/api/app/documents/${current.id}/share`, { method });
      if (method !== 'GET') await calendar.reloadDocument(current.id);
      if (method === 'DELETE') link = null;
    } catch (err) {
      shareMessage = errorMessage(err);
    }
  }

  async function shareIt() {
    if (!link?.url || !current) return;
    if (navigator.share) {
      try {
        await navigator.share({ title: current.title, url: link.url });
        return;
      } catch {
        // cancelled, or not allowed: copy instead
      }
    }
    try {
      await navigator.clipboard.writeText(link.url);
      shareMessage = 'Copied. Send it to your sitter.';
    } catch {
      shareMessage = 'Select the link and copy it.';
    }
  }
</script>

<Sheet
  title={editing ? (current ? 'Edit document' : 'New document') : (current?.title ?? '')}
  {onclose}
>
  {#if editing}
    <form class="flex flex-col gap-4 pt-1" onsubmit={save}>
      {#if !current}
        <div class="flex flex-wrap gap-2">
          {#each TEMPLATES as t (t.title)}
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-full border border-zinc-300 px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-800"
              onclick={() => useTemplate(t)}><Icon name={t.icon} class="size-4" />{t.title}</button
            >
          {/each}
        </div>
      {/if}
      <label>
        Name
        <!-- svelte-ignore a11y_autofocus -->
        <input
          bind:value={title}
          required
          maxlength="120"
          placeholder="Wolfy, The house, Grandma's place…"
          autofocus={!current}
        />
      </label>
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium">Icon</span>
        <div class="flex flex-wrap gap-1.5" role="radiogroup" aria-label="Icon">
          {#each ['doc', ...ICONS] as name (name)}
            <button
              type="button"
              role="radio"
              aria-checked={icon === name}
              aria-label={name}
              class="grid size-9 place-items-center rounded-xl border transition {icon === name
                ? 'border-brand-500 bg-brand-500/10 text-brand-700 dark:text-brand-400'
                : 'border-zinc-200 text-zinc-500 dark:border-zinc-700'}"
              onclick={() => (icon = name)}><Icon {name} class="size-4.5" /></button
            >
          {/each}
        </div>
      </div>
      <label>
        Notes
        <textarea
          bind:value={body}
          rows="9"
          maxlength="20000"
          placeholder="Food, walks, medicine, the vet's number, the door code…"></textarea>
      </label>

      <div class="flex flex-col gap-2">
        <span class="text-sm font-medium">Photos and files</span>
        {#if keptFiles.length || added.length}
          <ul class="flex flex-col gap-1.5">
            {#each keptFiles as name (name)}
              <li
                class="flex items-center gap-2 rounded-xl bg-zinc-50 px-3 py-2 text-sm dark:bg-zinc-800/60"
              >
                <Icon name={isImage(name) ? 'star' : 'clip'} class="size-4 text-zinc-400" />
                <span class="flex-1 truncate">{fileLabel(name)}</span>
                <button
                  type="button"
                  class="text-zinc-400 hover:text-red-600"
                  onclick={() => (removed = [...removed, name])}
                  aria-label="Remove {fileLabel(name)}"><Icon name="close" class="size-4" /></button
                >
              </li>
            {/each}
            {#each added as f, i (i)}
              <li class="flex items-center gap-2 rounded-xl bg-brand-500/5 px-3 py-2 text-sm">
                <Icon name="plus" class="size-4 text-brand-600" />
                <span class="flex-1 truncate">{f.name}</span>
                <button
                  type="button"
                  class="text-zinc-400 hover:text-red-600"
                  onclick={() => (added = added.filter((_, j) => j !== i))}
                  aria-label="Remove {f.name}"><Icon name="close" class="size-4" /></button
                >
              </li>
            {/each}
          </ul>
        {/if}
        <label class="btn-secondary cursor-pointer flex-row gap-1.5 self-start px-3 py-2 text-sm">
          <Icon name="clip" class="size-4" /> Add photos or files
          <input type="file" multiple class="sr-only" onchange={pickFiles} />
        </label>
      </div>

      {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}

      <div class="flex gap-2 pt-1">
        {#if canDelete}
          <button
            type="button"
            class="btn-secondary text-red-600 dark:text-red-400"
            onclick={remove}
            disabled={busy}
            aria-label="Delete"><Icon name="trash" /></button
          >
        {/if}
        {#if current}
          <button type="button" class="btn-secondary" onclick={() => (editing = false)}
            >Cancel</button
          >
        {/if}
        <button class="btn-primary flex-1" disabled={busy}
          >{busy ? 'Saving…' : current ? 'Save' : 'Add'}</button
        >
      </div>
    </form>
  {:else if current}
    {@const photos = current.files.filter(isImage)}
    {@const others = current.files.filter((f) => !isImage(f))}
    <div class="flex flex-col gap-4 pt-1">
      {#if current.body}
        <p class="text-[15px] leading-relaxed whitespace-pre-wrap">
          {#each linkParts(current.body) as part, i (i)}{#if part.href}<a
                class="font-medium text-brand-600 underline underline-offset-2 dark:text-brand-400"
                href={part.href}
                target={part.href.startsWith('http') ? '_blank' : undefined}
                rel="noopener noreferrer">{part.text}</a
              >{:else}{part.text}{/if}{/each}
        </p>
      {:else}
        <p class="text-sm text-zinc-500">Nothing written yet.</p>
      {/if}

      {#if photos.length}
        <div class="grid grid-cols-3 gap-2">
          {#each photos as name (name)}
            <a href={fileUrl(current, name)} target="_blank" rel="noopener noreferrer">
              <img
                src={fileUrl(current, name, '400x400')}
                alt={fileLabel(name)}
                class="aspect-square w-full rounded-xl object-cover"
                loading="lazy"
              />
            </a>
          {/each}
        </div>
      {/if}
      {#if others.length}
        <div class="flex flex-col gap-1.5">
          {#each others as name (name)}
            <a
              class="flex items-center gap-2 rounded-xl border border-zinc-200 px-3 py-2 text-sm hover:bg-zinc-50 dark:border-zinc-700 dark:hover:bg-zinc-800"
              href={fileUrl(current, name)}
              target="_blank"
              rel="noopener noreferrer"
              ><Icon name="clip" class="size-4 text-zinc-400" /><span class="truncate"
                >{fileLabel(name)}</span
              ></a
            >
          {/each}
        </div>
      {/if}

      <section class="flex flex-col gap-2 rounded-2xl bg-zinc-50 p-4 dark:bg-zinc-800/50">
        <h3 class="inline-flex items-center gap-2 text-sm font-semibold">
          <Icon name="share" class="size-4" /> Share with a sitter
        </h3>
        {#if current.shared && link?.url}
          <input
            readonly
            value={link.url}
            onfocus={(e) => e.currentTarget.select()}
            class="font-mono text-xs"
            aria-label="Private link"
          />
          <div class="flex flex-wrap gap-2">
            <button class="btn-primary gap-1.5 px-3 py-2 text-sm" onclick={shareIt}
              ><Icon name="send" class="size-4" />Send the link</button
            >
            <a
              class="btn-secondary px-3 py-2 text-sm"
              href={link.url}
              target="_blank"
              rel="noopener noreferrer">Preview <Icon name="external" class="ml-1 size-3.5" /></a
            >
            <button
              class="btn-secondary px-3 py-2 text-sm text-red-600 dark:text-red-400"
              onclick={() => getLink('DELETE')}>Stop sharing</button
            >
          </div>
          {#if !link.public}
            <p class="text-xs text-zinc-500">
              This link works on your tailnet only. For a sitter outside it, publish shared
              documents with Tailscale Funnel (CALIO_FEED_URL, see the deploy guide).
            </p>
          {/if}
        {:else}
          <p class="text-sm text-zinc-600 dark:text-zinc-400">
            Anyone with the private link can read this page and see its files, without an account.
            You can stop it working at any time.
          </p>
          <button class="btn-secondary self-start px-3 py-2 text-sm" onclick={() => getLink('POST')}
            >Make a private link</button
          >
        {/if}
        {#if shareMessage}<p class="text-sm text-zinc-500" role="status">{shareMessage}</p>{/if}
      </section>

      <button class="btn-secondary gap-1.5" onclick={edit}
        ><Icon name="edit" class="size-4" />Edit</button
      >
    </div>
  {/if}
</Sheet>
