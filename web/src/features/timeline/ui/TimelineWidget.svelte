<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import {
    type EventSeverity,
    type TimelineEvent,
    eventSources,
    groupByDay,
    timelineConfig,
  } from '../domain/timeline';
  import type { TimelineUseCases } from '../usecase/timelineUseCases';

  /** Props of the timeline widget (#74). */
  interface Props {
    config: Record<string, unknown>;
    timeline: TimelineUseCases;
  }
  let { config, timeline }: Props = $props();
  const poll = usePolling();

  const options = $derived(timelineConfig(config));
  let events = $state<TimelineEvent[]>([]);
  let source = $state('');
  let severity = $state<EventSeverity | ''>('');
  let error = $state('');

  $effect(() => {
    const filter = { source, severity, limit: options.limit };
    const load = () =>
      timeline
        .list(filter)
        .then((l) => ((events = l), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    return poll(load, 60_000);
  });

  const dayKey = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'short' });
  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
  const groups = $derived(groupByDay(events, dayKey));
</script>

{#if !options.compact}
  <div class="filters">
    <select class="input" bind:value={source} aria-label={t('timeline.source')}>
      <option value="">{t('timeline.all_sources')}</option>
      {#each eventSources as s (s)}<option value={s}>{t(`timeline.source.${s}`)}</option>{/each}
    </select>
    <select class="input" bind:value={severity} aria-label={t('timeline.severity')}>
      <option value="">{t('timeline.all_severities')}</option>
      <option value="info">{t('notifications.severity.info')}</option>
      <option value="warning">{t('notifications.severity.warning')}</option>
      <option value="critical">{t('notifications.severity.critical')}</option>
    </select>
  </div>
{/if}
{#if error}<p class="error">{error}</p>{/if}
<ol class="timeline">
  {#each groups as g (g.day)}
    <li class="day">{g.day}</li>
    {#each g.events as e (e.id)}
      <li class="event {e.severity}">
        <time datetime={e.at.toISOString()}>{time(e.at)}</time>
        <span class="msg">{e.message}</span>
      </li>
    {/each}
  {:else}
    <li class="muted">{t('timeline.empty')}</li>
  {/each}
</ol>

<style>
  .filters {
    display: flex;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .timeline {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .day {
    margin-top: var(--space-2);
    font-size: 0.8em;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--color-text-muted);
  }
  .event {
    display: flex;
    gap: var(--space-2);
    padding: 2px 0 2px var(--space-2);
    border-left: 3px solid var(--color-border, rgba(127, 127, 127, 0.3));
  }
  .event.warning {
    border-color: var(--color-warning);
  }
  .event.critical {
    border-color: var(--color-danger);
  }
  .event.info {
    border-color: var(--color-success);
  }
  time {
    font-variant-numeric: tabular-nums;
    color: var(--color-text-muted);
  }
</style>
