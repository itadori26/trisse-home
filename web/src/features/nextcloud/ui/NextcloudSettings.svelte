<script lang="ts">
  import { t } from '../../../shared/i18n';
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
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  {t('nextcloud.create_an_app_password_in_nextcloud')}
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>{t('nextcloud.nextcloud_url')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="https://cloud.example.com"
      required
    /></label
  >
  <label class="field"
    ><span>{t('nextcloud.username')}</span><input
      class="input"
      bind:value={username}
      autocomplete="off"
      required
    /></label
  >
  <label class="field">
    <span
      >{t('nextcloud.app_password', {
        value: hasPassword ? t('common.keep_stored') : '',
      })}</span
    >
    <input
      class="input"
      type="password"
      bind:value={appPassword}
      autocomplete="new-password"
      required={!hasPassword}
    />
  </label>
  <p class="muted">{t('nextcloud.the_app_password_is_stored_encrypted')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
