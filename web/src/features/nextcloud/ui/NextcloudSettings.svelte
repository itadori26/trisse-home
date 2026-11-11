<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { NextcloudUseCases } from '../usecase/nextcloudUseCases';

  /** Props of the Nextcloud connection settings. */
  interface Props {
    nextcloud: NextcloudUseCases;
  }
  let { nextcloud }: Props = $props();

  let url = $state('');
  let username = $state('');
  let appPassword = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    nextcloud.settings().then((s) => ((url = s.url), (username = s.username), (hasPassword = s.hasPassword)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await nextcloud.saveSettings(url, username, appPassword);
      hasPassword = s.hasPassword;
      appPassword = '';
      message = 'Connected and saved.';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  Create an app password in Nextcloud under Personal settings → Security → Devices &amp; sessions. Server
  details (free space, users, app updates) need an admin account and the serverinfo app; without them the
  widget shows the version and your quota.
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>Nextcloud URL</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="https://cloud.example.com"
      required
    /></label
  >
  <label class="field"
    ><span>Username</span><input class="input" bind:value={username} autocomplete="off" required /></label
  >
  <label class="field">
    <span>App password {hasPassword ? '(leave empty to keep the stored one)' : ''}</span>
    <input
      class="input"
      type="password"
      bind:value={appPassword}
      autocomplete="new-password"
      required={!hasPassword}
    />
  </label>
  <p class="muted">The app password is stored encrypted and never shown again.</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}>{busy ? 'Testing…' : 'Test & save'}</button>
</form>
