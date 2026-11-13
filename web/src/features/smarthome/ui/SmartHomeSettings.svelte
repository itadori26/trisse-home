<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { SmarthomeUseCases } from '../usecase/smarthomeUseCases';

  /** Props of the Home Assistant connection settings. */
  interface Props {
    smarthome: SmarthomeUseCases;
  }
  let { smarthome }: Props = $props();

  let url = $state('');
  let token = $state('');
  let hasToken = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    smarthome
      .settings()
      .then((s) => ((url = s.url), (hasToken = s.hasToken)))
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await smarthome.saveSettings(url, token);
      hasToken = s.hasToken;
      token = '';
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
    ><span>{t('smarthome.url')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://homeassistant.local:8123"
      required
    /></label
  >
  <label class="field">
    <span>{t('smarthome.token', { value: hasToken ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={token}
      autocomplete="new-password"
      required={!hasToken}
    />
  </label>
  <p class="muted">{t('smarthome.token_hint')}</p>
  <p class="muted">{t('smarthome.actions_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
