<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import { type HistoryPoint, type SystemSnapshot, daysUntilFull } from '../domain/system';
  import type { SystemUseCases } from '../usecase/systemUseCases';

  /** Props of the system widget. */
  interface Props {
    system: SystemUseCases;
  }
  let { system }: Props = $props();
  const poll = usePolling();

  /** The backend samples every 5 s. */
  const POLL_MS = 5000;
  let snap = $state<SystemSnapshot | null>(null);
  let history = $state<HistoryPoint[]>([]);
  /** History points are stored once a minute. */
  const HISTORY_POLL_MS = 60_000;
  let error = $state('');

  $effect(() => {
    const load = () =>
      system
        .snapshot()
        .then((s) => ((snap = s), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    const stopSnapshot = poll(load, POLL_MS);
    // The history is optional; failures still back off.
    const stopHistory = poll(() => system.history().then((h) => (history = h)), HISTORY_POLL_MS);
    return () => {
      stopSnapshot();
      stopHistory();
    };
  });

  const usage = (u: { used: number; total: number }) => `${formatBytes(u.used)} / ${formatBytes(u.total)}`;
</script>

{#if error && !snap}
  <p class="error">{error}</p>
{:else if snap}
  <div class="system">
    <Meter label={t('system.cpu_cores', { cpuCount: snap.cpuCount })} percent={snap.cpuPercent} />
    <Sparkline values={history.map((p) => p.cpu)} max={100} label={t('system.cpu_over_24_hours')} />
    <Meter label={t('system.memory')} percent={snap.memory.percent} detail={usage(snap.memory)} />
    <Sparkline values={history.map((p) => p.memory)} max={100} label={t('system.memory_over_24_hours')} />
    {#if snap.swap.total > 0}<Meter
        label={t('system.swap')}
        percent={snap.swap.percent}
        detail={usage(snap.swap)}
      />{/if}
    {#each snap.disks as d (d.label)}
      {@const full = daysUntilFull(d.forecast)}
      <Meter label={d.label} percent={d.usage.percent} detail={usage(d.usage)} />
      {#if full !== null}<p class="forecast muted" class:soon={full < 14}>
          {full === 0 ? t('system.forecast_full_now') : t('system.forecast_full_in', { days: full })}
        </p>{/if}
    {/each}
    {#if snap.disks.length}
      <Sparkline values={history.map((p) => p.disk)} max={100} label={t('system.disk_usage_over_24_hours')} />
    {/if}
    <p class="facts muted">
      <span title={t('system.load_average_1_5_15_min')}
        >{t('system.load', { value: snap.load.map((l) => l.toFixed(2)).join(' · ') })}</span
      >
      <span>{t('system.up', { value: formatUptime(snap.uptimeSeconds) })}</span>
      {#if snap.temperature > 0}<span>{snap.temperature.toFixed(0)} °C</span>{/if}
    </p>
  </div>
{:else}
  <p class="muted">{t('system.loading')}</p>
{/if}

<style>
  .forecast {
    margin: -0.25em 0 0.25em;
    font-size: 0.8em;
  }
  .forecast.soon {
    color: var(--color-warning);
  }
  /* Spread the meters over tall widgets instead of stacking them at the top. */
  .system {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-evenly;
  }
  .facts {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    margin: var(--space-2) 0 0;
    font-size: 0.8em;
  }
</style>
