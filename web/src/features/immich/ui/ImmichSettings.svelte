<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { ImmichUseCases } from '../usecase/immichUseCases';

  /** Props of the Immich connection settings. */
  interface Props {
    immich: ImmichUseCases;
  }
  let { immich }: Props = $props();

  let url = $state('');
  let apiKey = $state('');
  let hasApiKey = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    immich
      .settings()
      .then((s) => ((url = s.url), (hasApiKey = s.hasApiKey)))
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await immich.saveSettings(url, apiKey);
      hasApiKey = s.hasApiKey;
      apiKey = '';
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
    ><span>{t('immich.url')}</span><input
      class="input"
      type="url"
      bind:value={url}
      placeholder="http://192.168.1.2:2283"
      required
    /></label
  >
  <label class="field">
    <span>{t('immich.api_key', { value: hasApiKey ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={apiKey}
      autocomplete="new-password"
      required={!hasApiKey}
    />
  </label>
  <p class="muted">{t('immich.key_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
