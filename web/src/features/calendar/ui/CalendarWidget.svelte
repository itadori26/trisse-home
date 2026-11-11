<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { type AgendaDay, groupByDay, parseCalendarConfig } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';
  import MonthView from './MonthView.svelte';
  import QuickAddEvent from './QuickAddEvent.svelte';

  /** Props of the calendar widget. */
  interface Props {
    calendar: CalendarUseCases;
    config: Record<string, unknown>;
  }
  let { calendar, config }: Props = $props();

  const POLL_MS = 5 * 60 * 1000;
  const cfg = $derived(parseCalendarConfig(config));
  let days = $state<AgendaDay[]>([]);
  let loaded = $state(false);
  let warning = $state('');
  let error = $state('');
  let unconfigured = $state(false);
  let adding = $state(false);
  let reload = $state(0);

  $effect(() => {
    void reload;
    const c = cfg;
    const load = () =>
      calendar
        .events(c.days)
        .then((a) => {
          days = groupByDay(a.events, new Date(), c.days);
          warning = a.error;
          error = '';
          loaded = true;
        })
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
        });
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  const dayLabel = (d: Date) => {
    const today = new Date();
    const diff = Math.round(
      (d.getTime() - new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime()) / 86_400_000,
    );
    if (diff === 0) return t('time.today');
    if (diff === 1) return t('time.tomorrow');
    return d.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'short' });
  };
  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
</script>

{#if unconfigured}
  <p class="muted">{t('calendar.connect_a_caldav_calendar_under_settings')}</p>
{:else if error && !loaded}
  <p class="error">{error}</p>
{:else}
  <button
    class="btn btn-icon add"
    type="button"
    title={t('calendar.new_event')}
    aria-label={t('calendar.new_event')}
    onclick={() => (adding = true)}>+</button
  >
  {#if cfg.view === 'month'}
    <MonthView {calendar} {reload} />
  {:else}
    {#if warning}<p class="error small">{warning}</p>{/if}
    {#each days as day (day.key)}
      <section>
        <h4>{dayLabel(day.date)}</h4>
        <ul>
          {#each day.events as e (e.uid + e.start.getTime())}
            <li title={[e.location, e.calendar].filter(Boolean).join(' · ')}>
              <span class="time">{e.allDay ? t('calendar.all_day_short') : time(e.start)}</span>
              <span class="summary">{e.summary}</span>
            </li>
          {/each}
        </ul>
      </section>
    {:else}
      {#if loaded}<p class="muted">{t('calendar.no_events_in_the_next_days', { days: cfg.days })}</p>{/if}
    {/each}
  {/if}
{/if}
{#if adding}
  <QuickAddEvent {calendar} onclose={() => (adding = false)} oncreated={() => reload++} />
{/if}

<style>
  .add {
    float: right;
    margin: -4px 0 0 var(--space-1);
    padding: 0 6px;
    line-height: 1.3;
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: 0.8em;
    color: var(--color-text-muted);
  }
  section:first-child h4 {
    margin-top: 0;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    gap: var(--space-2);
    padding: 1px 0;
  }
  .time {
    flex: none;
    width: 4.5em;
    font-variant-numeric: tabular-nums;
    color: var(--color-accent);
    font-size: 0.85em;
  }
  .summary {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.75em;
  }
</style>
