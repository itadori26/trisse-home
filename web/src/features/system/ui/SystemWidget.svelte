<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import type { SystemSnapshot } from '../domain/system';
  import type { SystemUseCases } from '../usecase/systemUseCases';

  /** Props of the system widget. */
  interface Props {
    system: SystemUseCases;
  }
  let { system }: Props = $props();

  /** The backend samples every 5 s. */
  const POLL_MS = 5000;
  let snap = $state<SystemSnapshot | null>(null);
  let error = $state('');

  $effect(() => {
    const load = () =>
      system
        .snapshot()
        .then((s) => ((snap = s), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  const usage = (u: { used: number; total: number }) => `${formatBytes(u.used)} / ${formatBytes(u.total)}`;
</script>

{#if error && !snap}
  <p class="error">{error}</p>
{:else if snap}
  <div class="system">
    <Meter label="CPU ({snap.cpuCount} cores)" percent={snap.cpuPercent} />
    <Meter label="Memory" percent={snap.memory.percent} detail={usage(snap.memory)} />
    {#if snap.swap.total > 0}<Meter label="Swap" percent={snap.swap.percent} detail={usage(snap.swap)} />{/if}
    {#each snap.disks as d (d.label)}
      <Meter label={d.label} percent={d.usage.percent} detail={usage(d.usage)} />
    {/each}
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
