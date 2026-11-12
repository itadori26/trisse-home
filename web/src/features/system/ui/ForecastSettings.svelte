<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { SystemUseCases } from '../usecase/systemUseCases';

  /** Props of the disk forecast settings (#108). */
  interface Props {
    system: SystemUseCases;
  }
  let { system }: Props = $props();

  let windowDays = $state(14);
  let alertDays = $state(14);
  let message = $state('');
  let error = $state('');

  $effect(() => {
    system
      .forecastSettings()
      .then((s) => ((windowDays = s.windowDays), (alertDays = s.alertDays)))
      .catch((err) => (error = errorMessage(err)));
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    try {
      const s = await system.saveForecastSettings({ windowDays, alertDays });
      windowDays = s.windowDays;
      alertDays = s.alertDays;
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={save}>
  <p class="muted">{t('system.forecast_hint')}</p>
  <label class="field">
    <span>{t('system.forecast_window')}</span>
    <input class="input" type="number" min="3" max="90" bind:value={windowDays} />
  </label>
  <label class="field">
    <span>{t('system.forecast_alert')}</span>
    <input class="input" type="number" min="1" max="365" bind:value={alertDays} />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit">{t('common.save')}</button>
</form>
