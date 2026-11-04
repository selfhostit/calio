<script lang="ts">
  import { errorMessage, pb } from '../lib/pb';
  import Icon from './Icon.svelte';

  // The link calendar apps subscribe to. It is secret: anyone with it can read
  // the calendar, so it can be replaced with a new one.
  let link = $state<{ url: string; public: boolean } | null>(null);
  let busy = $state(false);
  let message = $state('');

  async function get(reset = false) {
    if (reset && !confirm('Make a new link? Calendars using the old one stop updating.')) return;
    busy = true;
    message = '';
    try {
      link = await pb.send('/api/app/feed-link' + (reset ? '/reset' : ''), {
        method: reset ? 'POST' : 'GET',
      });
      if (reset) message = 'New link made. Subscribe again with it.';
    } catch (err) {
      message = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function copy() {
    if (!link) return;
    try {
      await navigator.clipboard.writeText(link.url);
      message = 'Copied.';
    } catch {
      message = 'Select the link and copy it.';
    }
  }

  const webcal = $derived(link ? link.url.replace(/^https?:/, 'webcal:') : '');
  const google = $derived(
    link ? `https://calendar.google.com/calendar/r?cid=${encodeURIComponent(webcal)}` : '',
  );
</script>

<section class="card flex flex-col gap-3">
  <h3 class="font-semibold">Calendar feed</h3>
  <p class="text-sm text-zinc-600 dark:text-zinc-400">
    See Calio in Apple Calendar, Google Calendar or Outlook. They check for changes every hour or
    so; edits still happen here.
  </p>
  {#if link}
    <input
      readonly
      value={link.url}
      onfocus={(e) => e.currentTarget.select()}
      class="font-mono text-xs"
      aria-label="Feed link"
    />
    <div class="flex flex-wrap gap-2">
      <button class="btn-secondary px-3 py-2 text-sm" onclick={copy}>Copy link</button>
      <a class="btn-secondary px-3 py-2 text-sm" href={webcal}>Open in Calendar</a>
      {#if link.public}
        <a
          class="btn-secondary px-3 py-2 text-sm"
          href={google}
          target="_blank"
          rel="noopener noreferrer">Add to Google <Icon name="external" class="ml-1 size-3.5" /></a
        >
      {/if}
    </div>
    {#if !link.public}
      <p class="text-xs text-zinc-500">
        This link works on devices in your tailnet (Apple Calendar, Outlook, Thunderbird). Google
        Calendar fetches feeds from Google's own servers, so it needs the public feed address
        (CALIO_FEED_URL, through Tailscale Funnel).
      </p>
    {/if}
    <button
      class="self-start text-xs text-zinc-500 underline underline-offset-2"
      onclick={() => get(true)}
      disabled={busy}>Make a new link</button
    >
  {:else}
    <button class="btn-primary self-start" onclick={() => get()} disabled={busy}
      >{busy ? '…' : 'Get my feed link'}</button
    >
  {/if}
  {#if message}<p class="text-sm text-zinc-500" role="status">{message}</p>{/if}
</section>
