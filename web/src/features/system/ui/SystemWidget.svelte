<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import type { HistoryPoint, SystemSnapshot } from '../domain/system';
  import type { SystemUseCases } from '../usecase/systemUseCases';

  /** Props of the system widget. */
  interface Props {
    system: SystemUseCases;
  }
  let { system }: Props = $props();

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
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    const loadHistory = () =>
      system
        .history()
        .then((h) => (history = h))
        .catch(() => {});
    loadHistory();
    const historyTimer = setInterval(loadHistory, HISTORY_POLL_MS);
    return () => {
      clearInterval(timer);
      clearInterval(historyTimer);
    };
  });

  const usage = (u: { used: number; total: number }) => `${formatBytes(u.used)} / ${formatBytes(u.total)}`;
</script>

{#if error && !snap}
  <p class="error">{error}</p>
{:else if snap}
  <div class="system">
    <Meter label="CPU ({snap.cpuCount} cores)" percent={snap.cpuPercent} />
    <Sparkline values={history.map((p) => p.cpu)} max={100} label="CPU over 24 hours" />
    <Meter label="Memory" percent={snap.memory.percent} detail={usage(snap.memory)} />
    <Sparkline values={history.map((p) => p.memory)} max={100} label="Memory over 24 hours" />
    {#if snap.swap.total > 0}<Meter label="Swap" percent={snap.swap.percent} detail={usage(snap.swap)} />{/if}
    {#each snap.disks as d (d.label)}
      <Meter label={d.label} percent={d.usage.percent} detail={usage(d.usage)} />
    {/each}
    {#if snap.disks.length}
      <Sparkline values={history.map((p) => p.disk)} max={100} label="Disk usage over 24 hours" />
    {/if}
    <p class="facts muted">
      <span title="Load average 1/5/15 min">Load {snap.load.map((l) => l.toFixed(2)).join(' · ')}</span>
      <span>Up {formatUptime(snap.uptimeSeconds)}</span>
      {#if snap.temperature > 0}<span>{snap.temperature.toFixed(0)} °C</span>{/if}
    </p>
  </div>
{:else}
  <p class="muted">Loading…</p>
{/if}

<style>
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
