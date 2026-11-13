<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { calendarName, dateInputValue, type NewEventInput } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';

  /** Props of the quick-add dialog. */
  interface Props {
    calendar: CalendarUseCases;
    onclose: () => void;
    oncreated: () => void;
    /** Day to prefill (the selected day of the split view, #163); default today. */
    date?: Date;
  }
  let { calendar, onclose, oncreated, date = new Date() }: Props = $props();

  const pad = (n: number) => String(n).padStart(2, '0');
  const now = new Date();
  // svelte-ignore state_referenced_locally
  const initialDate = dateInputValue(date);
  const nextHour = (now.getHours() + 1) % 24;
  let calendars = $state<string[]>([]);
  let form = $state<NewEventInput>({
    calendarUrl: '',
    summary: '',
    location: '',
    allDay: false,
    date: initialDate,
    endDate: '',
    startTime: `${pad(nextHour)}:00`,
    endTime: `${pad((nextHour + 1) % 24)}:00`,
  });
  let error = $state('');
  let busy = $state(false);
  let loaded = $state(false);

  $effect(() => {
    calendar.settings().then((s) => {
      loaded = true;
      calendars = s.calendarUrls;
      form.calendarUrl ||= s.calendarUrls[0] ?? '';
    });
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    try {
      await calendar.createEvent(form);
      oncreated();
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('calendar.new_event')} {onclose}>
  {#if loaded && calendars.length === 0}
    <!-- ICS subscriptions are read-only; new events need a CalDAV calendar. -->
    <p class="muted" role="status">{t('calendar.quick_add_needs_caldav')}</p>
  {/if}
  <form onsubmit={submit}>
    <label class="field"
      ><span>{t('calendar.title')}</span><input
        class="input"
        bind:value={form.summary}
        maxlength="200"
        required
      /></label
    >
    {#if calendars.length > 1}
      <label class="field"
        ><span>{t('calendar.calendar')}</span><select class="input" bind:value={form.calendarUrl}>
          {#each calendars as url (url)}<option value={url}>{calendarName(url)}</option>{/each}
        </select></label
      >
    {/if}
    <label><input type="checkbox" bind:checked={form.allDay} /> {t('calendar.all_day')}</label>
    <div class="row">
      <label class="field"
        ><span>{form.allDay ? t('calendar.from') : t('calendar.date')}</span><input
          class="input"
          type="date"
          bind:value={form.date}
          required
        /></label
      >
      {#if form.allDay}
        <label class="field"
          ><span>{t('calendar.until_optional')}</span><input
            class="input"
            type="date"
            bind:value={form.endDate}
            min={form.date}
          /></label
        >
      {:else}
        <label class="field"
          ><span>{t('calendar.start')}</span><input
            class="input"
            type="time"
            bind:value={form.startTime}
            required
          /></label
        >
        <label class="field"
          ><span>{t('calendar.end')}</span><input
            class="input"
            type="time"
            bind:value={form.endTime}
            required
          /></label
        >
      {/if}
    </div>
    <label class="field"
      ><span>{t('calendar.location_optional')}</span><input
        class="input"
        bind:value={form.location}
        maxlength="200"
      /></label
    >
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}
        >{busy ? t('common.saving') : t('common.create')}</button
      >
      <button class="btn" type="button" onclick={onclose}>{t('calendar.cancel')}</button>
    </p>
  </form>
</Modal>

<style>
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row .field {
    flex: 1 1 8rem;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
</style>
