<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the change-password form. */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let current = $state('');
  let next = $state('');
  let confirmation = $state('');
  let message = $state('');
  let error = $state('');

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = message = '';
    try {
      await auth.changePassword(current, next, confirmation);
      current = next = confirmation = '';
      message = 'Password changed. All other sessions were signed out.';
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>Current password</span>
    <input class="input" type="password" autocomplete="current-password" bind:value={current} required />
  </label>
  <label class="field">
    <span>New password</span>
    <input class="input" type="password" autocomplete="new-password" bind:value={next} required />
  </label>
  <label class="field">
    <span>Repeat new password</span>
    <input class="input" type="password" autocomplete="new-password" bind:value={confirmation} required />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit">Change password</button>
</form>
