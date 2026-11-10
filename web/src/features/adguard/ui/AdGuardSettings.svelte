<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { AdGuardUseCases } from '../usecase/adguardUseCases';

  /** Props of the AdGuard connection settings. */
  interface Props {
    adguard: AdGuardUseCases;
  }
  let { adguard }: Props = $props();

  let url = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    adguard.settings().then((s) => ((url = s.url), (username = s.username), (hasPassword = s.hasPassword)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await adguard.saveSettings(url, username, password);
      hasPassword = s.hasPassword;
      password = '';
      message = 'Connected and saved.';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field"
    ><span>AdGuard Home URL</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.1.2:3000"
      required
    /></label
  >
  <label class="field"
    ><span>Username</span><input class="input" bind:value={username} autocomplete="off" required /></label
  >
  <label class="field">
    <span>Password {hasPassword ? '(leave empty to keep the stored one)' : ''}</span>
    <input
      class="input"
      type="password"
      bind:value={password}
      autocomplete="new-password"
      required={!hasPassword}
    />
  </label>
  <p class="muted">The password is stored encrypted and never shown again.</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}>{busy ? 'Testing…' : 'Test & save'}</button>
</form>
