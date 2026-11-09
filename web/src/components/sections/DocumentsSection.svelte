<script lang="ts">
  import type { Snippet } from 'svelte';
  import { calendar } from '../../lib/calendar.svelte';
  import type { Doc } from '../../lib/documents';
  import DocumentSheet from '../DocumentSheet.svelte';
  import Documents from '../Documents.svelte';
  import Fab from '../Fab.svelte';
  import Loaded from '../Loaded.svelte';

  // The household's notes and files.
  let { tabs }: { tabs: Snippet } = $props();

  let sheet = $state<{ doc: Doc | null } | null>(null);
</script>

<div class="flex flex-wrap items-center justify-between gap-3">
  {@render tabs()}
</div>

<Loaded
  loaded={calendar.loaded}
  error={calendar.error}
  onretry={() => ((calendar.error = ''), calendar.start())}
>
  <Documents onopen={(doc) => (sheet = { doc })} onadd={() => (sheet = { doc: null })} />
</Loaded>

<Fab label="Add a document" onclick={() => (sheet = { doc: null })} />

{#if sheet}
  <DocumentSheet doc={sheet.doc} onclose={() => (sheet = null)} />
{/if}
