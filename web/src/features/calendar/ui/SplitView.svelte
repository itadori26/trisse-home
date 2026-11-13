<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import {
    type CalendarEvent,
    dateInputValue,
    eventsByDay,
    eventsOnDay,
    monthGrid,
    upcoming,
    type WeekStart,
  } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';
  import DayEvents from './DayEvents.svelte';
  import MonthGrid from './MonthGrid.svelte';

  /**
   * Split view of the calendar widget (#163): month grid and events side by
   * side on wide widgets, stacked on narrow ones, only the list when the
   * widget is very low. Clicking a day shows its events, clicking it again
   * (or "Today") returns to the upcoming events.
   */
  interface Props {
    calendar: CalendarUseCases;
    weekStart: WeekStart;
    /** Increment to reload, e.g. after an event was created. */
    reload: number;
    /** No calendar is set up (404); the widget shows the hint instead. */
    onunconfigured: () => void;
    /** Opens the quick-add form, prefilled with the day. */
    onadd: (date: Date) => void;
  }
  let { calendar, weekStart, reload, onadd, onunconfigured }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 5 * 60 * 1000;
  const UPCOMING = 15;
  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth());
  let selected = $state<Date | null>(null);
  let events = $state<CalendarEvent[]>([]);
  let warning = $state('');
  let error = $state('');
  let open = $state<CalendarEvent | null>(null);

  const cells = $derived(monthGrid(year, month, new Date(), weekStart));
  const byDay = $derived(eventsByDay(events, cells));
  const selectedKey = $derived(selected ? dateInputValue(selected) : null);
  const title = $derived(
    new Date(year, month, 1).toLocaleDateString(intlLocale(), { month: 'long', year: 'numeric' }),
  );
  const isCurrentMonth = $derived(year === now.getFullYear() && month === now.getMonth());
  const list = $derived(
    selected
      ? eventsOnDay(events, selected)
      : upcoming(events, isCurrentMonth ? new Date() : new Date(year, month, 1), UPCOMING),
  );
  const heading = $derived(
    selected
      ? selected.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' })
      : t('calendar.upcoming'),
  );

  // The visible month, refreshed through the shared scheduler; the selection is kept.
  // Navigation reads the per-month cache; the periodic refresh and a new reload value renew it.
  let seenReload = 0;
  $effect(() => {
    const y = year;
    const m = month;
    const ws = weekStart;
    let first = reload === seenReload;
    seenReload = reload;
    const load = () => {
      const fresh = !first;
      first = false;
      return calendar
        .month(y, m, ws, fresh)
        .then((a) => ((events = a.events), (warning = a.error), (error = '')))
        .catch((err) => {
          if (err instanceof ApiError && err.status === 404) onunconfigured();
          error = errorMessage(err);
          throw err;
        });
    };
    return poll(load, POLL_MS);
  });

  function show(y: number, m: number, direction: number) {
    const d = new Date(y, m, 1);
    year = d.getFullYear();
    month = d.getMonth();
    // Prefetch the next month in the direction of travel (cached, no extra request later).
    if (direction) {
      const next = new Date(year, month + direction, 1);
      void calendar.month(next.getFullYear(), next.getMonth(), weekStart).catch(() => {});
    }
  }

  function select(date: Date) {
    // A second click on the selected day returns to the upcoming events.
    selected = selected && dateInputValue(selected) === dateInputValue(date) ? null : date;
    if (date.getMonth() !== month || date.getFullYear() !== year)
      show(date.getFullYear(), date.getMonth(), 0);
  }

  function move(date: Date) {
    selected = date;
    if (date.getMonth() !== month || date.getFullYear() !== year) {
      show(date.getFullYear(), date.getMonth(), date < new Date(year, month, 1) ? -1 : 1);
    }
  }

  function today() {
    selected = null;
    show(now.getFullYear(), now.getMonth(), 0);
  }
</script>

<div class="split">
  <div class="head">
    <button
      class="btn btn-icon"
      type="button"
      aria-label={t('calendar.previous_month')}
      onclick={() => show(year, month - 1, -1)}>‹</button
    >
    <strong class="title" aria-live="polite">{title}</strong>
    <button
      class="btn btn-icon"
      type="button"
      aria-label={t('calendar.next_month')}
      onclick={() => show(year, month + 1, 1)}>›</button
    >
    <button class="btn today" type="button" onclick={today}>{t('calendar.today_button')}</button>
    <button
      class="btn btn-icon add"
      type="button"
      aria-label={selected
        ? t('calendar.add_on', { date: selected.toLocaleDateString(intlLocale()) })
        : t('calendar.new_event')}
      title={t('calendar.new_event')}
      onclick={() => onadd(selected ?? new Date())}>+</button
    >
  </div>
  {#if error}<p class="error small">{error}</p>{:else if warning}<p class="error small">{warning}</p>{/if}
  <div class="body">
    <div class="month">
      <MonthGrid
        {cells}
        {byDay}
        selected={selectedKey}
        {weekStart}
        label={title}
        onselect={select}
        onmove={move}
      />
    </div>
    <div class="list">
      <DayEvents
        {heading}
        events={list}
        withDay={!selected}
        empty={selected ? t('calendar.no_events') : t('calendar.no_upcoming')}
        onopen={(e) => (open = e)}
      />
      {#if open}
        <!-- Loaded on first use: most views never open an event (#163 size budget). -->
        {#await import('./EventDetails.svelte') then { default: EventDetails }}
          <EventDetails event={open} onclose={() => (open = null)} />
        {/await}
      {/if}
    </div>
  </div>
</div>

<style>
  .split {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    height: 100%;
  }
  .head {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  .title {
    min-width: 9em;
    text-align: center;
  }
  .today {
    margin-left: auto;
    padding: 0 var(--space-2);
    font-size: 0.8em;
  }
  .add {
    padding: 0 6px;
  }
  .body {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-3);
    min-height: 0;
    flex: 1;
  }
  .list {
    position: relative;
    min-height: 0;
    overflow: auto;
  }
  /* Wide (about 2:1 and more): grid and events side by side. */
  @container widget (min-aspect-ratio: 2 / 1) and (min-width: 480px) {
    .body {
      grid-template-columns: minmax(14rem, 1.2fr) minmax(0, 1fr);
    }
  }
  /* Very low: only the events. */
  @container widget (max-height: 200px) {
    .month {
      display: none;
    }
  }
  .small {
    font-size: 0.75em;
  }
</style>
