<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { qrDataUrl } from '../../../shared/ui/qr';
  import {
    formatSecret,
    recoveryCodesText,
    type TwoFactorEnrollment,
    type TwoFactorStatus,
  } from '../domain/twoFactor';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the two-factor settings section. */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let status = $state<TwoFactorStatus | null>(null);
  let enrollment = $state<TwoFactorEnrollment | null>(null);
  let recoveryCodes = $state<string[] | null>(null);
  let code = $state('');
  let password = $state('');
  let error = $state('');

  async function run(action: () => Promise<void>) {
    error = '';
    try {
      await action();
      status = await auth.twoFactorStatus();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function download(codes: string[]) {
    const blob = new Blob([recoveryCodesText(codes, new Date())], { type: 'text/plain' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'trisse-home-recovery-codes.txt';
    a.click();
    URL.revokeObjectURL(a.href);
  }

  $effect(() => {
    run(async () => {});
  });
</script>

{#if recoveryCodes}
  <p>
    <strong>{t('auth.two_factor_authentication_is_on')}</strong>
    {t('auth.store_these_recovery_codes_in_a')}
  </p>
  <ul class="codes">
    {#each recoveryCodes as c (c)}<li>{c}</li>{/each}
  </ul>
  <p>
    <button class="btn" type="button" onclick={() => recoveryCodes && download(recoveryCodes)}
      >{t('auth.download_codes')}</button
    >
    <button class="btn btn-primary" type="button" onclick={() => (recoveryCodes = null)}
      >{t('auth.i_saved_them')}</button
    >
  </p>
{:else if status?.enabled}
  <p>
    {t('auth.two_factor_authentication_is')} <strong>{t('auth.on')}</strong>{t('auth.recovery_codes_left', {
      recoveryCodesLeft: status.recoveryCodesLeft,
    })}
  </p>
  <form
    class="row"
    onsubmit={(e) => {
      e.preventDefault();
      run(async () => {
        await auth.disableTwoFactor(password, code);
        password = code = '';
      });
    }}
  >
    <input
      class="input"
      type="password"
      autocomplete="current-password"
      placeholder={t('auth.password')}
      aria-label={t('auth.password')}
      bind:value={password}
      required
    />
    <input
      class="input"
      autocomplete="one-time-code"
      placeholder={t('auth.code')}
      aria-label={t('auth.authentication_code')}
      bind:value={code}
      required
    />
    <button class="btn btn-danger" type="submit">{t('auth.turn_off')}</button>
  </form>
{:else if enrollment}
  {@const e = enrollment}
  <p>
    {t('auth.scan_the_qr_code_with_an')}
  </p>
  <img
    class="qr"
    src={qrDataUrl(e.uri)}
    alt={t('auth.qr_code_for_the_authenticator_app')}
    width="200"
    height="200"
  />
  <p class="muted small">{t('auth.or_enter_the_key_manually')} <code>{formatSecret(e.secret)}</code></p>
  <form
    class="row"
    onsubmit={(ev) => {
      ev.preventDefault();
      run(async () => {
        recoveryCodes = await auth.confirmTwoFactor(code);
        enrollment = null;
        code = '';
      });
    }}
  >
    <input
      class="input"
      inputmode="numeric"
      autocomplete="one-time-code"
      placeholder="123456"
      aria-label={t('auth.code_from_the_app')}
      bind:value={code}
      required
    />
    <button class="btn btn-primary" type="submit">{t('auth.turn_on')}</button>
  </form>
{:else if status}
  <p class="muted">{t('auth.protect_your_login_with_a_second')}</p>
  <button
    class="btn"
    type="button"
    onclick={() =>
      run(async () => {
        enrollment = await auth.startTwoFactor();
      })}>{t('auth.set_up_two_factor_authentication')}</button
  >
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .row {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .row .input {
    flex: 1;
    min-width: 8rem;
  }
  .qr {
    image-rendering: pixelated;
    background: #fff;
    padding: 4px;
    border-radius: var(--radius-small);
  }
  .codes {
    columns: 2;
    font-family: var(--font-mono);
  }
  .small {
    font-size: 0.85rem;
  }
</style>
