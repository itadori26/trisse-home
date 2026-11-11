<script lang="ts">
  import { t } from '../../../shared/i18n';
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
  let username = $state('');
  /** Shows the name field for guest accounts. */
  let asGuest = $state(false);
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
      await auth.login(password, needsCode ? code : undefined, asGuest ? username : undefined);
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

<AuthCard title="trisse-home" subtitle={needsCode ? t('auth.enter_code_hint') : t('auth.please_log_in')}>
  <form onsubmit={submit}>
    {#if asGuest}
      <label class="field" class:hidden={needsCode}>
        <span>{t('auth.guest_name')}</span>
        <input
          class="input"
          name="username"
          autocomplete="username"
          autocapitalize="none"
          bind:value={username}
          required
        />
      </label>
    {/if}
    <label class="field" class:hidden={needsCode}>
      <span>{t('auth.password')}</span>
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
        <span>{t('auth.authentication_code_or_recovery_code')}</span>
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
    <button class="btn btn-primary" type="submit" disabled={busy}
      >{needsCode ? t('auth.verify') : t('auth.log_in')}</button
    >
    {#if !needsCode}
      <button class="btn btn-icon" type="button" onclick={() => ((asGuest = !asGuest), (error = ''))}
        >{asGuest ? t('auth.sign_in_as_owner') : t('auth.sign_in_as_guest')}</button
      >
    {/if}
    {#if needsCode}
      <button
        class="btn btn-icon"
        type="button"
        onclick={() => {
          needsCode = false;
          code = error = '';
        }}>{t('auth.back')}</button
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
