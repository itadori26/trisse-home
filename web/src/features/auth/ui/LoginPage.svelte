<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AuthCard from './AuthCard.svelte';

  /** Props of the login screen. */
  interface Props {
    auth: AuthUseCases;
    /** Called after a successful login. */
    ondone: () => void;
  }
  let { auth, ondone }: Props = $props();

  let password = $state('');
  let error = $state('');
  let busy = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = '';
    try {
      await auth.login(password);
      password = '';
      ondone();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<AuthCard title="trisse-home" subtitle="Please log in.">
  <form onsubmit={submit}>
    <label class="field">
      <span>Password</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class="input"
        type="password"
        name="password"
        autocomplete="current-password"
        bind:value={password}
        autofocus
        required
      />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>Log in</button>
  </form>
</AuthCard>
