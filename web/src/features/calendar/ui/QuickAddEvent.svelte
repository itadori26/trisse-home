<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { calendarName, type NewEventInput } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';

  /** Props of the quick-add dialog. */
  interface Props {
    calendar: CalendarUseCases;
    onclose: () => void;
    oncreated: () => void;
  }
  let { calendar, onclose, oncreated }: Props = $props();

  const pad = (n: number) => String(n).padStart(2, '0');
  const now = new Date();
  const nextHour = (now.getHours() + 1) % 24;
  let calendars = $state<string[]>([]);
  let form = $state<NewEventInput>({
    calendarUrl: '',
    summary: '',
    location: '',
    allDay: false,
    date: `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`,
    endDate: '',
    startTime: `${pad(nextHour)}:00`,
    endTime: `${pad((nextHour + 1) % 24)}:00`,
  });
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    calendar.settings().then((s) => {
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

<Modal title="New event" {onclose}>
  <form onsubmit={submit}>
    <label class="field"
      ><span>Title</span><input class="input" bind:value={form.summary} maxlength="200" required /></label
    >
    {#if calendars.length > 1}
      <label class="field"
        ><span>Calendar</span><select class="input" bind:value={form.calendarUrl}>
          {#each calendars as url (url)}<option value={url}>{calendarName(url)}</option>{/each}
        </select></label
      >
    {/if}
    <label><input type="checkbox" bind:checked={form.allDay} /> All day</label>
    <div class="row">
      <label class="field"
        ><span>{form.allDay ? 'From' : 'Date'}</span><input
          class="input"
          type="date"
          bind:value={form.date}
          required
        /></label
      >
      {#if form.allDay}
        <label class="field"
          ><span>Until (optional)</span><input
            class="input"
            type="date"
            bind:value={form.endDate}
            min={form.date}
          /></label
        >
      {:else}
        <label class="field"
          ><span>Start</span><input class="input" type="time" bind:value={form.startTime} required /></label
        >
        <label class="field"
          ><span>End</span><input class="input" type="time" bind:value={form.endTime} required /></label
        >
      {/if}
    </div>
    <label class="field"
      ><span>Location (optional)</span><input
        class="input"
        bind:value={form.location}
        maxlength="200"
      /></label
    >
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}>{busy ? 'Saving…' : 'Create'}</button>
      <button class="btn" type="button" onclick={onclose}>Cancel</button>
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
