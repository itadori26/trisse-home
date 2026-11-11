<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type CalendarEvent, eventsOnDay, monthGrid } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';

  /** Props of the month view. */
  interface Props {
    calendar: CalendarUseCases;
    /** Increment to reload, e.g. after an event was created. */
    reload: number;
  }
  let { calendar, reload }: Props = $props();

  const today = new Date();
  let year = $state(today.getFullYear());
  let month = $state(today.getMonth());
  let events = $state<CalendarEvent[]>([]);
  let error = $state('');
  let selected = $state<Date | null>(null);

  const cells = $derived(monthGrid(year, month, today));
  const title = $derived(
    new Date(year, month, 1).toLocaleDateString(intlLocale(), { month: 'long', year: 'numeric' }),
  );
  const weekdays = Array.from({ length: 7 }, (_, i) =>
    new Date(2024, 0, 1 + i).toLocaleDateString(intlLocale(), { weekday: 'narrow' }),
  );

  $effect(() => {
    void reload;
    const first = cells[0]?.date;
    const last = cells[cells.length - 1]?.date;
    if (!first || !last) return;
    const to = new Date(last.getFullYear(), last.getMonth(), last.getDate() + 1);
    calendar
      .range(first, to)
      .then((a) => ((events = a.events), (error = a.error)))
      .catch((err) => (error = errorMessage(err)));
  });

  function shift(delta: number) {
    const d = new Date(year, month + delta, 1);
    year = d.getFullYear();
    month = d.getMonth();
    selected = null;
  }
  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
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
    onclick={() => ((year = today.getFullYear()), (month = today.getMonth()))}>{title}</button
  >
  <button class="btn btn-icon" type="button" aria-label={t('calendar.next_month')} onclick={() => shift(1)}
    >›</button
  >
</div>
{#if error}<p class="error small">{error}</p>{/if}
<div class="grid" role="grid" aria-label={title}>
  {#each weekdays as w, i (i)}<span class="wd" role="columnheader">{w}</span>{/each}
  {#each cells as c (c.key)}
    {@const dayEvents = eventsOnDay(events, c.date)}
    <button
      type="button"
      class="day"
      class:out={!c.inMonth}
      class:today={c.isToday}
      class:sel={selected?.getTime() === c.date.getTime()}
      aria-label={t('calendar.events', {
        value: c.date.toLocaleDateString(intlLocale()),
        length: dayEvents.length,
      })}
      onclick={() => (selected = c.date)}
    >
      <span class="num">{c.date.getDate()}</span>
      <span class="items">
        {#each dayEvents.slice(0, 3) as e, i (i)}<span class="item" title={e.summary}
            ><span class="text">{e.summary}</span></span
          >{/each}
      </span>
    </button>
  {/each}
</div>
{#if selected}
  {@const list = eventsOnDay(events, selected)}
  <div class="detail">
    <strong
      >{selected.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' })}</strong
    >
    {#each list as e (e.uid + e.start.getTime())}
      <div class="ev">
        <span class="time">{e.allDay ? t('calendar.all_day_short') : time(e.start)}</span>
        {e.summary}
      </div>
    {:else}
      <div class="muted">{t('calendar.no_events')}</div>
    {/each}
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
  .grid {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 2px;
  }
  .wd {
    text-align: center;
    font-size: 0.7em;
    color: var(--color-text-muted);
  }
  .day {
    min-height: clamp(1.8em, 7cqh, 6em);
    padding: 2px;
    border: 1px solid transparent;
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    color: inherit;
    font: inherit;
    font-size: 0.8em;
    text-align: left;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    cursor: pointer;
  }
  .out {
    opacity: 0.45;
  }
  .today .num {
    color: var(--color-accent);
    font-weight: 700;
  }
  .sel {
    border-color: var(--color-accent);
  }
  .items {
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    min-width: 0;
  }
  /* Small cells: dots; larger cells (container query) show short titles. */
  .item {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--color-accent);
  }
  .text {
    display: none;
  }
  @container widget (min-width: 420px) and (min-height: 320px) {
    .items {
      flex-direction: column;
    }
    .item {
      width: auto;
      height: auto;
      border-radius: 3px;
      background: color-mix(in srgb, var(--color-accent) 20%, transparent);
      padding: 0 3px;
      font-size: 0.85em;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .text {
      display: inline;
    }
  }
  .detail {
    margin-top: var(--space-2);
    font-size: 0.85em;
  }
  .ev {
    display: flex;
    gap: var(--space-2);
  }
  .time {
    color: var(--color-accent);
    width: 4.5em;
    flex: none;
  }
  .small {
    font-size: 0.75em;
  }
</style>
