<script lang="ts">
  import { t } from '../../../shared/i18n';
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
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  {t('fritzbox.connection_state_and_throughput_work_wit')}
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>{t('fritzbox.fritz_box_address_tr_064_port')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.178.1:49000"
    /></label
  >
  <label class="field"
    ><span>{t('fritzbox.username_optional')}</span><input
      class="input"
      bind:value={username}
      autocomplete="off"
    /></label
  >
  <label class="field">
    <span
      >{t('fritzbox.password', {
        value: hasPassword ? t('common.keep_stored') : t('common.optional'),
      })}</span
    >
    <input class="input" type="password" bind:value={password} autocomplete="new-password" />
  </label>
  <p class="muted">{t('fritzbox.the_password_is_stored_encrypted_and')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
