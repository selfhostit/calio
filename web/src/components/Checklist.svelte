<script lang="ts">
  import type { ChecklistItem } from '../lib/entries';
  import Icon from './Icon.svelte';

  // A list of things to do or pack. onchange runs after every tick, add and
  // removal, so an entry that already exists can save each change right away.
  let {
    items = $bindable(),
    label,
    suggestions = [],
    onchange,
  }: {
    items: ChecklistItem[];
    label: string;
    suggestions?: { label: string; items: string[] }[];
    onchange?: () => void;
  } = $props();

  let text = $state('');
  const done = $derived(items.filter((i) => i.done).length);

  function add() {
    const t = text.trim();
    if (!t) return;
    items = [...items, { text: t, done: false }];
    text = '';
    onchange?.();
  }

  function toggle(i: number) {
    items = items.map((item, j) => (j === i ? { ...item, done: !item.done } : item));
    onchange?.();
  }

  function remove(i: number) {
    items = items.filter((_, j) => j !== i);
    onchange?.();
  }

  function use(list: string[]) {
    items = list.map((t) => ({ text: t, done: false }));
    onchange?.();
  }
</script>

<div class="flex flex-col gap-2">
  <div class="flex items-center justify-between">
    <span class="text-sm font-medium">{label}</span>
    {#if items.length}<span class="text-xs text-zinc-500">{done} of {items.length} done</span>{/if}
  </div>
  {#if items.length}
    <ul
      class="flex flex-col divide-y divide-zinc-100 rounded-xl border border-zinc-200 dark:divide-zinc-800 dark:border-zinc-700"
    >
      {#each items as item, i (i)}
        <li class="flex items-center gap-3 px-3 py-2">
          <input
            type="checkbox"
            checked={item.done}
            onchange={() => toggle(i)}
            aria-label={item.text}
          />
          <span class="flex-1 text-sm {item.done ? 'text-zinc-400 line-through' : ''}"
            >{item.text}</span
          >
          <button
            type="button"
            class="grid size-7 place-items-center rounded-full text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800"
            onclick={() => remove(i)}
            aria-label="Remove {item.text}"><Icon name="close" class="size-3.5" /></button
          >
        </li>
      {/each}
    </ul>
  {:else if suggestions.length}
    <div class="flex flex-wrap gap-2">
      {#each suggestions as s (s.label)}
        <button
          type="button"
          class="rounded-full border border-zinc-300 px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-800"
          onclick={() => use(s.items)}>{s.label}</button
        >
      {/each}
    </div>
  {/if}
  <div class="flex gap-2">
    <input
      bind:value={text}
      placeholder="Add an item"
      maxlength="200"
      onkeydown={(e) => {
        if (e.key === 'Enter') {
          e.preventDefault();
          add();
        }
      }}
    />
    <button type="button" class="btn-secondary shrink-0 px-3" onclick={add} aria-label="Add item"
      ><Icon name="plus" /></button
    >
  </div>
</div>
