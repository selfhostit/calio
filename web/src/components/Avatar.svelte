<script lang="ts">
  import { avatarColor, firstName } from '../lib/calendar.svelte';
  import type { User } from '../lib/pb';

  let { person, size = 'size-7 text-xs' }: { person: User | undefined; size?: string } = $props();
</script>

{#if person}
  <span
    class="inline-flex shrink-0 items-center justify-center rounded-full font-semibold ring-2 ring-white dark:ring-zinc-900 {avatarColor(
      person.id,
    )} {size}"
    title={person.name || person.email || firstName(person)}
  >
    {firstName(person).slice(0, 1).toUpperCase()}
  </span>
{/if}
