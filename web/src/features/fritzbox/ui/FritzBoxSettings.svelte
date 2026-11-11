<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { FritzBoxUseCases } from '../usecase/fritzboxUseCases';

  /** Props of the FRITZ!Box connection settings. */
  interface Props {
    fritzbox: FritzBoxUseCases;
  }
  let { fritzbox }: Props = $props();

  let url = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    fritzbox.settings().then((s) => ((url = s.url), (username = s.username), (hasPassword = s.hasPassword)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await fritzbox.saveSettings(url, username, password);
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

<p class="muted">
  Connection state and throughput work without a login if “Transmit status information over UPnP” is enabled
  on the box. For the call list and link speed, create a dedicated FRITZ!Box user (System → FRITZ!Box Users)
  with only the rights “FRITZ!Box settings” and “Voice messages, faxes, FRITZ!App Fon and call list”.
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>FRITZ!Box address (TR-064 port)</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.178.1:49000"
    /></label
  >
  <label class="field"
    ><span>Username (optional)</span><input class="input" bind:value={username} autocomplete="off" /></label
  >
  <label class="field">
    <span>Password {hasPassword ? '(leave empty to keep the stored one)' : '(optional)'}</span>
    <input class="input" type="password" bind:value={password} autocomplete="new-password" />
  </label>
  <p class="muted">The password is stored encrypted and never shown again.</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}>{busy ? 'Testing…' : 'Test & save'}</button>
</form>
