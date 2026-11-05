<script lang="ts">
  import { calendar } from '../lib/calendar.svelte';
  import { BEHAVIOURS, type Category } from '../lib/categories';
  import type { Kind } from '../lib/entries';
  import { COLORS, ICONS, LOOKS, type Color } from '../lib/look';
  import { errorMessage } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Icon from './Icon.svelte';

  // The categories the household uses: switch them on and off, change their
  // name, color and icon, or add new ones. Entries in a category that's
  // switched off stay on the calendar; it just isn't offered for new ones.

  // the category being changed ('new' for one being added)
  let open = $state<string | null>(null);
  let name = $state('');
  let color = $state<Color>('sky');
  let icon = $state('star');
  let kind = $state<Kind>('event');
  let busy = $state(false);
  let error = $state('');

  const categories = $derived(calendar.sortedCategories);
  const used = (c: Category) => calendar.entries.filter((e) => e.category === c.id).length;

  function edit(c: Category | null) {
    error = '';
    if (open === (c?.id ?? 'new')) {
      open = null;
      return;
    }
    open = c?.id ?? 'new';
    name = c?.name ?? '';
    // a new one gets a color nobody uses yet
    color = c?.color ?? COLORS.find((x) => !categories.some((c) => c.color === x)) ?? 'sky';
    icon = c?.icon ?? 'star';
    kind = c?.kind ?? 'event';
  }

  async function toggle(c: Category) {
    try {
      await calendar.saveCategory(c.id, { enabled: !c.enabled });
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    const data: Partial<Category> = { name: name.trim(), color, icon, kind };
    try {
      if (open === 'new') {
        const last = categories.at(-1)?.sort ?? 0;
        await calendar.saveCategory(null, { ...data, enabled: true, sort: last + 1 });
      } else if (open) {
        await calendar.saveCategory(open, data);
      }
      open = null;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove(c: Category) {
    const n = used(c);
    const what = n
      ? `Delete ${c.name}? Its ${n} ${n === 1 ? 'entry stays' : 'entries stay'} on the calendar, in plain colors.`
      : `Delete ${c.name}?`;
    if (!confirm(what)) return;
    try {
      await calendar.removeCategory(c.id);
      open = null;
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#snippet editor(c: Category | null)}
  <form
    class="flex flex-col gap-3 border-t border-zinc-100 px-3 pt-3 pb-3 dark:border-zinc-800"
    onsubmit={save}
  >
    <label>
      Name
      <!-- svelte-ignore a11y_autofocus -->
      <input
        bind:value={name}
        required
        maxlength="40"
        placeholder="House sitting, Football, Grandma…"
        autofocus={!c}
      />
    </label>
    <div class="flex flex-col gap-1.5">
      <span class="text-sm font-medium">Color</span>
      <div class="flex flex-wrap gap-2" role="radiogroup" aria-label="Color">
        {#each COLORS as x (x)}
          <button
            type="button"
            role="radio"
            aria-checked={color === x}
            aria-label={x}
            class="size-8 rounded-full {LOOKS[x]
              .dot} ring-offset-2 transition ring-offset-white dark:ring-offset-zinc-900 {color ===
            x
              ? 'ring-2 ring-zinc-900 dark:ring-zinc-100'
              : ''}"
            onclick={() => (color = x)}
          ></button>
        {/each}
      </div>
    </div>
    <div class="flex flex-col gap-1.5">
      <span class="text-sm font-medium">Icon</span>
      <div class="flex flex-wrap gap-1.5" role="radiogroup" aria-label="Icon">
        {#each ICONS as x (x)}
          <button
            type="button"
            role="radio"
            aria-checked={icon === x}
            aria-label={x}
            class="grid size-9 place-items-center rounded-xl border transition {icon === x
              ? LOOKS[color].chip
              : 'border-zinc-200 text-zinc-500 dark:border-zinc-700'}"
            onclick={() => (icon = x)}><Icon name={x} class="size-4.5" /></button
          >
        {/each}
      </div>
    </div>
    <label>
      Works like
      <select bind:value={kind}>
        {#each BEHAVIOURS as b (b.kind)}<option value={b.kind}>{b.label}</option>{/each}
      </select>
      <span class="text-xs font-normal text-zinc-500"
        >{BEHAVIOURS.find((b) => b.kind === kind)?.hint}</span
      >
    </label>
    {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}
    <div class="flex gap-2">
      {#if c && (c.created_by === session.user?.id || session.isOwner)}
        <button
          type="button"
          class="btn-secondary px-3 py-2 text-red-600 dark:text-red-400"
          onclick={() => remove(c)}
          aria-label="Delete {c.name}"><Icon name="trash" class="size-4" /></button
        >
      {/if}
      <button type="button" class="btn-secondary px-3 py-2 text-sm" onclick={() => (open = null)}
        >Cancel</button
      >
      <button class="btn-primary flex-1 px-3 py-2 text-sm" disabled={busy}
        >{c ? 'Save' : 'Add'}</button
      >
    </div>
  </form>
{/snippet}

<section class="card flex flex-col gap-3">
  <div>
    <h3 class="font-semibold">Categories</h3>
    <p class="text-sm text-zinc-600 dark:text-zinc-400">
      Choose what goes on your calendar. Switch off what you don't need.
    </p>
  </div>
  <ul class="flex flex-col gap-2">
    {#each categories as c (c.id)}
      {@const look = LOOKS[c.color] ?? LOOKS.indigo}
      <li class="rounded-2xl border border-zinc-200 dark:border-zinc-700">
        <div class="flex items-center gap-3 p-2">
          <button
            class="flex min-w-0 flex-1 items-center gap-3 rounded-xl text-left"
            onclick={() => edit(c)}
            aria-expanded={open === c.id}
          >
            <span
              class="grid size-9 shrink-0 place-items-center rounded-xl {look.tile} {c.enabled
                ? ''
                : 'opacity-50'}"><Icon name={c.icon} class="size-4.5" /></span
            >
            <span class="min-w-0 flex-1">
              <span class="block truncate font-medium {c.enabled ? '' : 'text-zinc-400'}"
                >{c.name}</span
              >
              <span class="block text-xs text-zinc-500"
                >{BEHAVIOURS.find((b) => b.kind === c.kind)?.label}</span
              >
            </span>
          </button>
          <input
            type="checkbox"
            role="switch"
            class="switch"
            checked={c.enabled}
            onchange={() => toggle(c)}
            aria-label="Use {c.name}"
          />
        </div>
        {#if open === c.id}{@render editor(c)}{/if}
      </li>
    {/each}
    {#if open === 'new'}
      <li class="rounded-2xl border border-zinc-200 dark:border-zinc-700">
        {@render editor(null)}
      </li>
    {/if}
  </ul>
  {#if open !== 'new'}
    <button class="btn-secondary gap-1.5 self-start px-3 py-2 text-sm" onclick={() => edit(null)}
      ><Icon name="plus" class="size-4" />Add a category</button
    >
  {/if}
  {#if error && !open}<p class="text-sm text-red-600 dark:text-red-400" role="alert">
      {error}
    </p>{/if}
</section>
