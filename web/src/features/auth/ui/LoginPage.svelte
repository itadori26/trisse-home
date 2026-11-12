<script lang="ts">
  import { untrack } from 'svelte';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { ssoErrorKey } from '../domain/auth';
  import type { SecondFactorChallenge } from '../domain/passkey';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AuthCard from './AuthCard.svelte';

  /** Props of the login screen. */
  interface Props {
    auth: AuthUseCases;
    /** Called after a successful login. */
    ondone: () => void;
  }
  let { auth, ondone }: Props = $props();

  const params = new URLSearchParams(location.search);
  // The ticket is read once when the page opens; later changes come from submit().
  let ticket = $state(params.get('sso_ticket') ?? untrack(() => auth.ssoTicket));
  const ssoError = ssoErrorKey(params.get('sso_error'));
  if (params.has('sso_ticket') || params.has('sso_error')) history.replaceState(null, '', location.pathname);
  let password = $state('');
  let username = $state('');
  /** Shows the name field for guest accounts. */
  let asGuest = $state(false);
  let code = $state('');
  /** True once the server asked for the second factor. */
  let needsCode = $state(false);
  /** Passkey challenge of the second login step. */
  let challenge = $state<SecondFactorChallenge | null>(null);

  async function usePasskey() {
    if (!challenge) return;
    busy = true;
    error = '';
    try {
      await auth.loginWithPasskey(challenge);
      password = code = '';
      ondone();
    } catch (err) {
      // The ticket is single use: start again with the password.
      error = errorMessage(err);
      challenge = null;
      needsCode = false;
    } finally {
      busy = false;
    }
  }
  const passwordHidden = $derived(needsCode || !!ticket || (!auth.sso.localLogin && !asGuest));
  let error = $state('');
  let busy = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = '';
    try {
      if (ticket) {
        await auth.completeSso(ticket, code);
        ondone();
        return;
      }
      const second = await auth.login(password, needsCode ? code : undefined, asGuest ? username : undefined);
      if (second) {
        challenge = second;
        needsCode = true;
        return;
      }
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
  {#if ssoError}<p class="error" role="alert">{t(ssoError)}</p>{/if}
  {#if auth.sso.oidc && !ticket && !needsCode}
    <a class="btn btn-primary sso" href="/api/auth/oidc/start">{t('auth.sign_in_with_sso')}</a>
  {/if}
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
    <label class="field" class:hidden={passwordHidden}>
      <span>{t('auth.password')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class="input"
        type="password"
        name="password"
        autocomplete="current-password"
        bind:value={password}
        autofocus
        required={!passwordHidden}
      />
    </label>
    {#if challenge}
      <button class="btn btn-primary" type="button" disabled={busy} onclick={usePasskey}
        >{t('auth.use_passkey')}</button
      >
      <p class="muted">{t('auth.or_enter_code')}</p>
    {/if}
    {#if needsCode || ticket}
      <label class="field">
        <span>{t('auth.authentication_code_or_recovery_code')}</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          class="input code"
          name="code"
          inputmode="text"
          autocomplete="one-time-code"
          bind:value={code}
          autofocus={!challenge}
          required
        />
      </label>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if !passwordHidden || needsCode || ticket}
      <button class="btn btn-primary" type="submit" disabled={busy}
        >{needsCode || ticket ? t('auth.verify') : t('auth.log_in')}</button
      >
    {/if}
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
