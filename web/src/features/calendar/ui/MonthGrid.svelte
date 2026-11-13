<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import {
    type CalendarEvent,
    dotsForDay,
    type MonthCell,
    moveDay,
    type WeekStart,
    weekdayOrder,
  } from '../domain/calendar';

  /**
   * The month grid (#163), presentational: the parent owns the month, the
   * events and the selection. Keyboard: arrows move by a day or a week,
   * PageUp/PageDown by a month, Home/End within the week, Enter or Space
   * selects. One cell is in the tab order at a time (roving tabindex).
   */
  interface Props {
    cells: MonthCell[];
    /** Events per day key (eventsByDay). */
    byDay: Map<string, CalendarEvent[]>;
    /** Key of the selected day, or null. */
    selected: string | null;
    weekStart: WeekStart;
    /** Accessible name of the grid, e.g. "November 2026". */
    label: string;
    /** Short event titles in large cells instead of dots (the month view). */
    titles?: boolean;
    /** A click or Enter/Space on a day. */
    onselect: (date: Date) => void;
    /** A keyboard move; the parent selects the day and changes the month if needed. */
    onmove: (date: Date) => void;
  }
  let { cells, byDay, selected, weekStart, label, titles = false, onselect, onmove }: Props = $props();

  let grid = $state<HTMLElement>();
  let moved = false;

  const weekdays = $derived(
    weekdayOrder(weekStart).map((d) => ({
      short: new Date(2024, 0, 7 + d).toLocaleDateString(intlLocale(), { weekday: 'narrow' }),
      long: new Date(2024, 0, 7 + d).toLocaleDateString(intlLocale(), { weekday: 'long' }),
    })),
  );
  const rows = $derived(Array.from({ length: cells.length / 7 }, (_, i) => cells.slice(i * 7, i * 7 + 7)));
  // The cell in the tab order: the selection, else today, else the first day of the month.
  const focusKey = $derived(
    selected && cells.some((c) => c.key === selected)
      ? selected
      : (cells.find((c) => c.isToday)?.key ?? cells.find((c) => c.inMonth)?.key ?? cells[0]?.key),
  );

  // After a keyboard move the focus follows the new selection.
  $effect(() => {
    const key = focusKey;
    if (!moved || !grid || !key) return;
    moved = false;
    grid.querySelector<HTMLElement>(`[data-key="${key}"]`)?.focus();
  });

  function cellLabel(c: MonthCell, n: number): string {
    const date = c.date.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' });
    if (n === 0) return t('calendar.day_no_events', { date });
    return n === 1 ? t('calendar.day_one_event', { date }) : t('calendar.day_events', { date, n });
  }

  function onkeydown(e: KeyboardEvent, c: MonthCell) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onselect(c.date);
      return;
    }
    const next = moveDay(c.date, e.key, weekStart);
    if (!next) return;
    e.preventDefault();
    moved = true;
    onmove(next);
  }
</script>

<div class="grid" role="grid" aria-label={label} bind:this={grid}>
  <div class="row" role="row">
    {#each weekdays as w (w.long)}<span class="wd" role="columnheader" aria-label={w.long}>{w.short}</span
      >{/each}
  </div>
  {#each rows as row (row[0]?.key)}
    <div class="row" role="row">
      {#each row as c (c.key)}
        {@const events = byDay.get(c.key) ?? []}
        {@const dots = dotsForDay(events)}
        <div
          role="gridcell"
          tabindex={c.key === focusKey ? 0 : -1}
          data-key={c.key}
          class="day"
          class:out={!c.inMonth}
          class:today={c.isToday}
          class:sel={c.key === selected}
          aria-selected={c.key === selected}
          aria-current={c.isToday ? 'date' : undefined}
          aria-label={cellLabel(c, events.length)}
          onclick={() => onselect(c.date)}
          onkeydown={(e) => onkeydown(e, c)}
        >
          <span class="num" aria-hidden="true">{c.date.getDate()}</span>
          {#if titles}
            <span class="titles" aria-hidden="true">
              {#each events.slice(0, 3) as e, i (i)}<span class="title" style:--event-color={e.color || null}
                  >{e.summary}</span
                >{/each}
            </span>
          {/if}
          <span class="dots" class:with-titles={titles} aria-hidden="true">
            {#each dots.colors as color, i (i)}<span class="dot" style:--event-color={color || null}
              ></span>{/each}
            {#if dots.more}<span class="more">+{dots.more}</span>{/if}
          </span>
        </div>
      {/each}
    </div>
  {/each}
</div>

<style>
  .grid {
    display: grid;
    gap: 2px;
  }
  .row {
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
    min-height: clamp(1.9em, 6cqh, 6em);
    padding: 2px 3px;
    border: 1px solid transparent;
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    font-size: 0.8em;
    display: flex;
    flex-direction: column;
    gap: 1px;
    overflow: hidden;
    cursor: pointer;
  }
  .day:focus-visible {
    outline: 2px solid var(--color-accent);
    outline-offset: 1px;
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
    background: color-mix(in srgb, var(--color-accent) 12%, var(--color-surface-2));
  }
  .dots {
    display: flex;
    align-items: center;
    gap: 2px;
    min-height: 6px;
  }
  .dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--event-color, var(--color-accent));
  }
  .more {
    font-size: 0.75em;
    line-height: 1;
    color: var(--color-text-muted);
  }
  .titles {
    display: none;
  }
  /* The month view shows short titles in large cells. */
  @container widget (min-width: 420px) and (min-height: 320px) {
    .titles {
      display: flex;
      flex-direction: column;
      gap: 1px;
      min-width: 0;
    }
    .title {
      border-radius: 3px;
      background: color-mix(in srgb, var(--event-color, var(--color-accent)) 20%, transparent);
      padding: 0 3px;
      font-size: 0.85em;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .dots.with-titles {
      display: none;
    }
  }
</style>
