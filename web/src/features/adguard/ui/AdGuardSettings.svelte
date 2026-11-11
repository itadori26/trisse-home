<script lang="ts">
  import { t } from '../../../shared/i18n';
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
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field"
    ><span>{t('adguard.adguard_home_url')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.1.2:3000"
      required
    /></label
  >
  <label class="field"
    ><span>{t('adguard.username')}</span><input
      class="input"
      bind:value={username}
      autocomplete="off"
      required
    /></label
  >
  <label class="field">
    <span>{t('adguard.password', { value: hasPassword ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={password}
      autocomplete="new-password"
      required={!hasPassword}
    />
  </label>
  <p class="muted">{t('adguard.the_password_is_stored_encrypted_and')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
