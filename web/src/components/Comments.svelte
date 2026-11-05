<script lang="ts">
  import { onMount } from 'svelte';
  import { calendar, firstName } from '../lib/calendar.svelte';
  import { ago, type Comment } from '../lib/comments';
  import type { Entry } from '../lib/entries';
  import { errorMessage, pb } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Avatar from './Avatar.svelte';
  import Icon from './Icon.svelte';

  // The conversation on an entry, live. A new comment notifies everyone else.
  let { entry }: { entry: Entry } = $props();

  let comments = $state<Comment[]>([]);
  let text = $state('');
  let busy = $state(false);
  let error = $state('');
  let now = $state(new Date());

  const sorted = $derived([...comments].sort((a, b) => a.created.localeCompare(b.created)));

  onMount(() => {
    const id = entry.id;
    const c = pb.collection<Comment>('comments');
    const filter = pb.filter('entry = {:id}', { id });
    c.getFullList({ filter, sort: 'created' })
      .then((list) => (comments = list))
      .catch((err) => (error = errorMessage(err)));
    const stop = c.subscribe(
      '*',
      ({ action, record }) => {
        const rest = comments.filter((x) => x.id !== record.id);
        comments = action === 'delete' ? rest : [...rest, record];
      },
      { filter },
    );
    const timer = setInterval(() => (now = new Date()), 30_000);
    return () => {
      stop.then((unsubscribe) => unsubscribe());
      clearInterval(timer);
    };
  });

  async function send(e: SubmitEvent) {
    e.preventDefault();
    const t = text.trim();
    if (!t) return;
    busy = true;
    error = '';
    try {
      const saved = await pb.collection<Comment>('comments').create({ entry: entry.id, text: t });
      comments = [...comments.filter((x) => x.id !== saved.id), saved];
      text = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove(c: Comment) {
    if (!confirm('Delete this comment?')) return;
    try {
      await pb.collection('comments').delete(c.id);
      comments = comments.filter((x) => x.id !== c.id);
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<section class="mt-6 flex flex-col gap-3 border-t border-zinc-100 pt-4 dark:border-zinc-800">
  <h3 class="inline-flex items-center gap-2 text-sm font-semibold">
    <Icon name="chat" class="size-4" /> Comments
    {#if sorted.length}<span class="text-zinc-400">{sorted.length}</span>{/if}
  </h3>

  {#each sorted as c (c.id)}
    {@const author = calendar.person(c.author)}
    {@const mine = c.author === session.user?.id}
    <article class="group flex gap-2.5">
      <Avatar person={author} />
      <div class="min-w-0 flex-1">
        <p class="flex items-baseline gap-2 text-xs text-zinc-500">
          <span class="font-semibold text-zinc-800 dark:text-zinc-200"
            >{mine ? 'You' : firstName(author) || 'Someone'}</span
          >
          <span>{ago(c.created, now)}</span>
          {#if mine || session.isOwner}
            <button
              class="ml-auto text-zinc-400 opacity-0 group-hover:opacity-100 focus:opacity-100 hover:text-red-600"
              onclick={() => remove(c)}
              aria-label="Delete comment"><Icon name="trash" class="size-3.5" /></button
            >
          {/if}
        </p>
        <p
          class="mt-0.5 inline-block rounded-2xl rounded-tl-md bg-zinc-100 px-3 py-2 text-sm whitespace-pre-line dark:bg-zinc-800"
        >
          {c.text}
        </p>
      </div>
    </article>
  {/each}

  <form class="flex items-end gap-2" onsubmit={send}>
    <textarea
      bind:value={text}
      rows="1"
      maxlength="2000"
      placeholder="Write a comment"
      class="min-h-10 flex-1 resize-none"
      aria-label="Comment"
      onkeydown={(e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
          e.preventDefault();
          e.currentTarget.form?.requestSubmit();
        }
      }}></textarea>
    <button class="btn-primary shrink-0 px-3" disabled={busy || !text.trim()} aria-label="Send"
      ><Icon name="send" /></button
    >
  </form>
  {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}
</section>
