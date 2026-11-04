<script lang="ts">
  import { untrack } from 'svelte';
  import { calendar } from '../lib/calendar.svelte';
  import { ROLES, type Contact, type Role } from '../lib/contacts';
  import { errorMessage } from '../lib/pb';
  import { session } from '../lib/session.svelte';
  import Icon from './Icon.svelte';
  import Sheet from './Sheet.svelte';

  let { contact = null, onclose }: { contact?: Contact | null; onclose: () => void } = $props();

  const editing = untrack(() => contact);
  let name = $state(editing?.name ?? '');
  let role = $state<Role>(editing?.role ?? 'vet');
  let phone = $state(editing?.phone ?? '');
  let email = $state(editing?.email ?? '');
  let address = $state(editing?.address ?? '');
  let notes = $state(editing?.notes ?? '');
  let busy = $state(false);
  let error = $state('');

  const canDelete = $derived(
    !!editing && (editing.created_by === session.user?.id || session.isOwner),
  );

  async function save(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    try {
      await calendar.saveContact(editing?.id ?? null, {
        name: name.trim(),
        role,
        phone: phone.trim(),
        email: email.trim(),
        address: address.trim(),
        notes: notes.trim(),
      });
      onclose();
    } catch (err) {
      const fields = (err as { response?: { data?: Record<string, { message: string }> } }).response
        ?.data;
      error = (fields && Object.values(fields)[0]?.message) || errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove() {
    if (!editing || !confirm(`Delete ${editing.name}?`)) return;
    busy = true;
    try {
      await calendar.removeContact(editing.id);
      onclose();
    } catch (err) {
      error = errorMessage(err);
      busy = false;
    }
  }
</script>

<Sheet title={editing ? 'Edit contact' : 'Add a contact'} {onclose}>
  <form class="flex flex-col gap-4 pt-1" onsubmit={save}>
    <label>
      Name
      <!-- svelte-ignore a11y_autofocus -->
      <input
        bind:value={name}
        required
        maxlength="120"
        placeholder="Dr. Papadopoulou"
        autofocus={!editing}
      />
    </label>
    <label>
      Role
      <select bind:value={role}>
        {#each ROLES as r (r.role)}<option value={r.role}>{r.label}</option>{/each}
      </select>
    </label>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <label
        >Phone <input type="tel" bind:value={phone} maxlength="40" placeholder="Optional" /></label
      >
      <label>Email <input type="email" bind:value={email} placeholder="Optional" /></label>
    </div>
    <label>
      Address
      <input bind:value={address} maxlength="200" placeholder="Address, “lat, lon” or a map link" />
    </label>
    <label>
      Notes
      <textarea
        bind:value={notes}
        rows="4"
        maxlength="2000"
        placeholder="Opening hours, prices, feeding routine, door code…"></textarea>
    </label>

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
      <button class="btn-primary flex-1" disabled={busy}
        >{busy ? 'Saving…' : editing ? 'Save' : 'Add'}</button
      >
    </div>
  </form>
</Sheet>
