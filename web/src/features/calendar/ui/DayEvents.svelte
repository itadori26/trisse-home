<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import type { CalendarEvent } from '../domain/calendar';

  /** The events next to the month grid (#163): one day, or the upcoming ones. */
  interface Props {
    heading: string;
    events: CalendarEvent[];
    /** Show the day on each row (the upcoming list spans several days). */
    withDay?: boolean;
    empty: string;
    onopen: (event: CalendarEvent) => void;
  }
  let { heading, events, withDay = false, empty, onopen }: Props = $props();

  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
  const day = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'short' });
</script>

<section class="events" aria-label={heading}>
  <h4>{heading}</h4>
  {#if events.length}
    <ul>
      {#each events as e (e.uid + e.start.getTime())}
        <li style:--event-color={e.color || null}>
          <button type="button" class="ev" onclick={() => onopen(e)}>
            <span class="when">
              {#if withDay}<span class="day">{day(e.start)}</span>{/if}
              {e.allDay ? t('calendar.all_day_short') : time(e.start)}
            </span>
            <span class="what">
              <span class="summary">{e.summary}</span>
              {#if e.location}<span class="muted where">{e.location}</span>{/if}
            </span>
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="muted">{empty}</p>
  {/if}
</section>

<style>
  h4 {
    margin: 0 0 var(--space-1);
    font-size: 0.8em;
    color: var(--color-text-muted);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    border-left: 3px solid var(--event-color, var(--color-accent));
    margin-bottom: 2px;
  }
  .ev {
    display: flex;
    gap: var(--space-2);
    width: 100%;
    padding: 2px var(--space-1);
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
    border-radius: var(--radius-small);
  }
  .ev:hover,
  .ev:focus-visible {
    background: var(--color-surface-2);
  }
  .when {
    flex: none;
    width: 5.5em;
    font-size: 0.85em;
    color: var(--color-accent);
    font-variant-numeric: tabular-nums;
    display: flex;
    flex-direction: column;
  }
  .day {
    color: var(--color-text-muted);
  }
  .what {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .summary,
  .where {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .where {
    font-size: 0.8em;
  }
</style>
