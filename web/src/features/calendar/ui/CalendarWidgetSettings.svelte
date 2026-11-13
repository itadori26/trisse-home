<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseCalendarConfig } from '../domain/calendar';

  /** Props of the calendar widget settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let days = $state(parseCalendarConfig(config).days);
  // svelte-ignore state_referenced_locally
  let view = $state(parseCalendarConfig(config).view);
  // svelte-ignore state_referenced_locally
  let weekStart = $state(parseCalendarConfig(config).weekStart);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ days, view, weekStart });
  }}
>
  <label class="field"
    ><span>{t('calendar.view')}</span><select class="input" bind:value={view}>
      <option value="agenda">{t('calendar.agenda')}</option>
      <option value="month">{t('calendar.month')}</option>
      <option value="split">{t('calendar.split')}</option>
    </select></label
  >
  {#if view !== 'agenda'}
    <label class="field"
      ><span>{t('calendar.week_start')}</span><select class="input" bind:value={weekStart}>
        <option value={1}>{t('calendar.monday')}</option>
        <option value={0}>{t('calendar.sunday')}</option>
      </select></label
    >
  {/if}
  {#if view === 'agenda'}
    <label class="field"
      ><span>{t('calendar.days_ahead', { days: days })}</span><input
        type="range"
        min="1"
        max="31"
        bind:value={days}
      /></label
    >
  {/if}
  <button class="btn btn-primary" type="submit">{t('calendar.save')}</button>
</form>
