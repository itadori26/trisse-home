<script lang="ts">
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
      message = 'Saved.';
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<p class="muted">
  DHL tracking needs a free API key: register at
  <a href="https://developer.dhl.com" target="_blank" rel="noopener noreferrer">developer.dhl.com</a>, create
  an app with the “Shipment Tracking – Unified” API and paste its key here. The free plan allows 250 requests
  a day; parcels are refreshed hourly. Tracking numbers stay on your server.
</p>
<form onsubmit={submit}>
  <label class="field">
    <span>DHL API key {hasKey ? '(leave empty to keep the stored one)' : ''}</span>
    <input class="input" type="password" bind:value={apiKey} autocomplete="off" />
  </label>
  <label class="field">
    <span>Hide delivered parcels after {hideDays} {hideDays === 1 ? 'day' : 'days'}</span>
    <input type="range" min="0" max="30" bind:value={hideDays} />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit">Save</button>
</form>
