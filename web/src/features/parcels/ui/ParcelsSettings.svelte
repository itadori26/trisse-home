<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { ParcelsUseCases } from '../usecase/parcelsUseCases';

  /** Props of the parcel tracking settings. */
  interface Props {
    parcels: ParcelsUseCases;
  }
  let { parcels }: Props = $props();

  let apiKey = $state('');
  let hasKey = $state(false);
  let hideDays = $state(3);
  let error = $state('');
  let message = $state('');

  $effect(() => {
    parcels.settings().then((s) => ((hasKey = s.hasDhlApiKey), (hideDays = s.hideDeliveredDays)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    try {
      const s = await parcels.saveSettings(apiKey, hideDays);
      hasKey = s.hasDhlApiKey;
      apiKey = '';
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<p class="muted">
  {t('parcels.dhl_tracking_needs_a_free_api')}
  <a href="https://developer.dhl.com" target="_blank" rel="noopener noreferrer">developer.dhl.com</a>{t(
    'parcels.create_an_app_with_the_shipment',
  )}
</p>
<form onsubmit={submit}>
  <label class="field">
    <span>{t('parcels.dhl_api_key', { value: hasKey ? t('common.keep_stored') : '' })}</span>
    <input class="input" type="password" bind:value={apiKey} autocomplete="off" />
  </label>
  <label class="field">
    <span>{t(hideDays === 1 ? 'parcels.hide_after_day' : 'parcels.hide_after_days', { n: hideDays })}</span>
    <input type="range" min="0" max="30" bind:value={hideDays} />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit">{t('parcels.save')}</button>
</form>
