<script lang="ts">
  import Icon from './Icon.svelte';

  // The round add button in the corner; each section says what it adds.
  let { label, onclick }: { label: string; onclick: () => void } = $props();
</script>

<button
  class="fixed right-5 bottom-[max(1.25rem,env(safe-area-inset-bottom))] z-30 grid size-14 place-items-center rounded-full bg-brand-600 text-white shadow-lg shadow-zinc-900/20 transition hover:bg-brand-700 active:scale-95 sm:right-[max(1.25rem,calc(50vw-24rem+1.25rem))]"
  {onclick}
  aria-label={label}
>
  <Icon name="plus" class="size-6" />
</button>
