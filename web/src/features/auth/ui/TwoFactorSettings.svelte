<script lang="ts">
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
    <strong>Two-factor authentication is on.</strong> Store these recovery codes in a safe place — each works once
    if you lose your phone. They are shown only now.
  </p>
  <ul class="codes">
    {#each recoveryCodes as c (c)}<li>{c}</li>{/each}
  </ul>
  <p>
    <button class="btn" type="button" onclick={() => recoveryCodes && download(recoveryCodes)}
      >Download codes</button
    >
    <button class="btn btn-primary" type="button" onclick={() => (recoveryCodes = null)}>I saved them</button>
  </p>
{:else if status?.enabled}
  <p>Two-factor authentication is <strong>on</strong>. Recovery codes left: {status.recoveryCodesLeft}.</p>
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
      placeholder="Password"
      aria-label="Password"
      bind:value={password}
      required
    />
    <input
      class="input"
      autocomplete="one-time-code"
      placeholder="Code"
      aria-label="Authentication code"
      bind:value={code}
      required
    />
    <button class="btn btn-danger" type="submit">Turn off</button>
  </form>
{:else if enrollment}
  {@const e = enrollment}
  <p>
    Scan the QR code with an authenticator app (e.g. Aegis, Google Authenticator, 1Password), then enter the
    code it shows.
  </p>
  <img class="qr" src={qrDataUrl(e.uri)} alt="QR code for the authenticator app" width="200" height="200" />
  <p class="muted small">Or enter the key manually: <code>{formatSecret(e.secret)}</code></p>
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
      aria-label="Code from the app"
      bind:value={code}
      required
    />
    <button class="btn btn-primary" type="submit">Turn on</button>
  </form>
{:else if status}
  <p class="muted">Protect your login with a second factor from an authenticator app.</p>
  <button
    class="btn"
    type="button"
    onclick={() =>
      run(async () => {
        enrollment = await auth.startTwoFactor();
      })}>Set up two-factor authentication</button
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
