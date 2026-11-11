<script lang="ts">
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
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
  let code = $state('');
  /** True once the server asked for the second factor. */
  let needsCode = $state(false);
  let error = $state('');
  let busy = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = '';
    try {
      await auth.login(password, needsCode ? code : undefined);
      password = code = '';
      ondone();
    } catch (err) {
      if (err instanceof ApiError && err.field === 'totpCode' && !needsCode) {
        needsCode = true;
      } else {
        error = errorMessage(err);
      }
    } finally {
      busy = false;
    }
  }
</script>

<AuthCard
  title="trisse-home"
  subtitle={needsCode ? 'Enter the code from your authenticator app.' : 'Please log in.'}
>
  <form onsubmit={submit}>
    <label class="field" class:hidden={needsCode}>
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
    {#if needsCode}
      <label class="field">
        <span>Authentication code (or recovery code)</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          class="input code"
          name="code"
          inputmode="text"
          autocomplete="one-time-code"
          bind:value={code}
          autofocus
          required
        />
      </label>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{needsCode ? 'Verify' : 'Log in'}</button>
    {#if needsCode}
      <button
        class="btn btn-icon"
        type="button"
        onclick={() => {
          needsCode = false;
          code = error = '';
        }}>Back</button
      >
    {/if}
  </form>
</AuthCard>

<style>
  .hidden {
    display: none;
  }
  .code {
    font-family: var(--font-mono);
    letter-spacing: 0.15em;
    font-size: 1.2rem;
  }
</style>
