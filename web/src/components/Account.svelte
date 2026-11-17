<script lang="ts">
  import { errorMessage, pb } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import { toasts } from '../lib/toast.svelte';
  import Avatar from './Avatar.svelte';

  // Who you are, signing out, and changing your own password, at the top of Settings.
  const user = $derived(session.user!);

  let open = $state(false);
  let current = $state('');
  let password = $state('');
  let again = $state('');
  let busy = $state(false);
  let error = $state('');

  async function change(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    if (password !== again) {
      error = "The new passwords don't match.";
      return;
    }
    busy = true;
    try {
      await pb.collection('users').update(user.id, {
        oldPassword: current,
        password,
        passwordConfirm: again,
      });
    } catch (err) {
      const data = (err as { response?: { data?: Record<string, unknown> } }).response?.data;
      error = data?.oldPassword
        ? 'Your current password is wrong.'
        : data?.password
          ? 'The new password needs at least 8 characters.'
          : errorMessage(err);
      busy = false;
      return;
    }
    // a new password signs out every device, this one too: sign in again with it
    try {
      await session.login(user.email!, password);
      current = password = again = '';
      open = false;
      toasts.show('Your password is changed. Other devices need to sign in again.');
    } catch {
      pb.authStore.clear();
    } finally {
      busy = false;
    }
  }
</script>

<section class="card flex flex-col gap-4">
  <div class="flex items-center gap-3">
    <Avatar person={user} size="size-11 text-base" />
    <div class="min-w-0 flex-1">
      <p class="truncate font-semibold">{user.name || user.email}</p>
      <p class="truncate text-sm text-zinc-500">
        {user.email}{session.isOwner ? ' · owner' : ''}
      </p>
    </div>
    <button class="btn-secondary px-3 py-2 text-sm" onclick={() => session.logout()}
      >Sign out</button
    >
  </div>
  {#if open}
    <form class="flex flex-col gap-4" onsubmit={change}>
      <label>
        Current password
        <input type="password" autocomplete="current-password" bind:value={current} required />
      </label>
      <label>
        New password
        <input
          type="password"
          autocomplete="new-password"
          bind:value={password}
          required
          minlength="8"
          placeholder="At least 8 characters"
        />
      </label>
      <label>
        New password again
        <input
          type="password"
          autocomplete="new-password"
          bind:value={again}
          required
          minlength="8"
        />
      </label>
      {#if session.isOwner}
        <p class="text-sm text-zinc-500">This doesn't change your password for the admin page.</p>
      {/if}
      {#if error}<p class="text-sm text-red-600 dark:text-red-400" role="alert">{error}</p>{/if}
      <div class="flex gap-2">
        <button class="btn-primary flex-1" disabled={busy}
          >{busy ? 'Saving…' : 'Change password'}</button
        >
        <button
          type="button"
          class="btn-secondary"
          onclick={() => ((open = false), (error = ''), (current = password = again = ''))}
          >Cancel</button
        >
      </div>
    </form>
  {:else}
    <button class="btn-secondary self-start px-3 py-2 text-sm" onclick={() => (open = true)}
      >Change password</button
    >
  {/if}
</section>
