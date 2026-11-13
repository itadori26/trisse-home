<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type ClockConfig, isValidTimeZone, parseClockConfig } from '../domain/clock';

  /** Props of the clock settings form. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<ClockConfig>(parseClockConfig(config));
  let error = $state('');

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!isValidTimeZone(form.timeZone.trim())) {
      error = t('clock.unknown_time_zone');
      return;
    }
    await onsave({ ...form, timeZone: form.timeZone.trim() });
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>{t('clock.label')}</span>
    <input class="input" bind:value={form.label} maxlength="40" />
  </label>
  <label class="field">
    <span>{t('clock.time_zone_empty_this_device')}</span>
    <input class="input" bind:value={form.timeZone} placeholder="Europe/Berlin" />
  </label>
  <label><input type="checkbox" bind:checked={form.hour12} /> {t('clock.12_hour_clock')}</label><br />
  <label><input type="checkbox" bind:checked={form.showSeconds} /> {t('clock.show_seconds')}</label><br />
  <label><input type="checkbox" bind:checked={form.showDate} /> {t('clock.show_date')}</label><br />
  <label><input type="checkbox" bind:checked={form.showWeek} /> {t('clock.show_week')}</label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p><button class="btn btn-primary" type="submit">{t('clock.save')}</button></p>
</form>
