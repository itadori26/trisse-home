<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import {
    type CalendarEvent,
    dateInputValue,
    eventsByDay,
    eventsOnDay,
    monthGrid,
    type WeekStart,
  } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';
  import DayEvents from './DayEvents.svelte';
  import MonthGrid from './MonthGrid.svelte';

  /** The month view: the grid with short titles, the selected day below it. */
  interface Props {
    calendar: CalendarUseCases;
    weekStart: WeekStart;
    /** Increment to reload, e.g. after an event was created. */
    reload: number;
    /** No calendar is set up (404); the widget shows the hint instead. */
    onunconfigured: () => void;
  }
  let { calendar, weekStart, reload, onunconfigured }: Props = $props();

  const today = new Date();
  let year = $state(today.getFullYear());
  let month = $state(today.getMonth());
  let events = $state<CalendarEvent[]>([]);
  let error = $state('');
  let selected = $state<Date | null>(null);
  let open = $state<CalendarEvent | null>(null);

  const cells = $derived(monthGrid(year, month, today, weekStart));
  const byDay = $derived(eventsByDay(events, cells));
  const title = $derived(
    new Date(year, month, 1).toLocaleDateString(intlLocale(), { month: 'long', year: 'numeric' }),
  );

  // Navigation reads the per-month cache; a new reload value (an event was created) renews it.
  let seenReload = 0;
  $effect(() => {
    const fresh = reload !== seenReload;
    seenReload = reload;
    calendar
      .month(year, month, weekStart, fresh)
      .then((a) => ((events = a.events), (error = a.error)))
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) onunconfigured();
        error = errorMessage(err);
      });
  });

  function show(date: Date) {
    year = date.getFullYear();
    month = date.getMonth();
  }
  function shift(delta: number) {
    show(new Date(year, month + delta, 1));
    selected = null;
  }
  function select(date: Date) {
    selected = date;
    if (date.getMonth() !== month || date.getFullYear() !== year) show(date);
  }
</script>

<div class="head">
  <button
    class="btn btn-icon"
    type="button"
    aria-label={t('calendar.previous_month')}
    onclick={() => shift(-1)}>‹</button
  >
  <button
    class="title"
    type="button"
    title={t('calendar.back_to_today')}
    onclick={() => ((selected = null), show(today))}>{title}</button
  >
  <button class="btn btn-icon" type="button" aria-label={t('calendar.next_month')} onclick={() => shift(1)}
    >›</button
  >
</div>
{#if error}<p class="error small">{error}</p>{/if}
<MonthGrid
  {cells}
  {byDay}
  selected={selected ? dateInputValue(selected) : null}
  {weekStart}
  label={title}
  titles
  onselect={select}
  onmove={select}
/>
{#if selected}
  <div class="detail">
    <DayEvents
      heading={selected.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' })}
      events={eventsOnDay(events, selected)}
      empty={t('calendar.no_events')}
      onopen={(e) => (open = e)}
    />
    {#if open}
      <!-- Loaded on first use: most views never open an event (#163 size budget). -->
      {#await import('./EventDetails.svelte') then { default: EventDetails }}
        <EventDetails event={open} onclose={() => (open = null)} />
      {/await}
    {/if}
  </div>
{/if}

<style>
  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-1);
  }
  .title {
    background: none;
    border: 0;
    color: inherit;
    font: inherit;
    font-weight: 600;
    cursor: pointer;
  }
  .detail {
    position: relative;
    margin-top: var(--space-2);
    font-size: 0.9em;
  }
  .small {
    font-size: 0.75em;
  }
</style>
