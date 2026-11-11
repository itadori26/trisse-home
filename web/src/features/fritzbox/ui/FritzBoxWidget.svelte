<script lang="ts">
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import { callIcons, callParty, type FritzBoxStatus, formatRate } from '../domain/fritzbox';
  import type { FritzBoxUseCases } from '../usecase/fritzboxUseCases';

  /** Props of the FRITZ!Box widget. */
  interface Props {
    fritzbox: FritzBoxUseCases;
  }
  let { fritzbox }: Props = $props();

  const POLL_MS = 15_000;
  let status = $state<FritzBoxStatus | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  $effect(() => {
    const load = () =>
      fritzbox
        .status()
        .then((s) => ((status = s), (error = ''), (unconfigured = false)))
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
        });
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  const when = (d: Date) =>
    d.toDateString() === new Date().toDateString()
      ? d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : d.toLocaleDateString([], { day: '2-digit', month: '2-digit' });
</script>

{#if unconfigured}
  <p class="muted">Connect your FRITZ!Box under Settings → FRITZ!Box.</p>
{:else if error && !status}
  <p class="error">{error}</p>
{:else if status}
  {@const wan = status.wan}
  <p class="head">
    <span class="dot" class:on={wan.connected} aria-hidden="true"></span>
    <strong>{wan.connected ? 'Online' : 'Offline'}</strong>
    {#if wan.connected && wan.uptimeSeconds}<span class="muted small"
        >for {formatUptime(wan.uptimeSeconds)}</span
      >{/if}
    {#if wan.externalIp}<span class="muted small ip" title="External IP">{wan.externalIp}</span>{/if}
  </p>
  {#if wan.throughputKnown}
    {#if wan.downstreamMaxBits > 0}
      <Meter label="↓ {formatRate(wan.downstreamBytesPerSecond)}" percent={wan.downstreamPercent} />
      <Meter label="↑ {formatRate(wan.upstreamBytesPerSecond)}" percent={wan.upstreamPercent} />
    {:else}
      <p class="rates">
        <span>↓ <strong>{formatRate(wan.downstreamBytesPerSecond)}</strong></span>
        <span>↑ <strong>{formatRate(wan.upstreamBytesPerSecond)}</strong></span>
      </p>
    {/if}
  {/if}
  {#if status.calls.length}
    <ul class="calls">
      {#each status.calls as c, i (i)}
        <li
          class:missed={c.type === 'missed'}
          title="{c.number}{c.durationMinutes ? ` · ${c.durationMinutes} min` : ''}"
        >
          <span class="icon" aria-label={c.type}>{callIcons[c.type]}</span>
          <span class="party">{callParty(c)}</span>
          <span class="muted small">{when(c.at)}</span>
        </li>
      {/each}
    </ul>
  {/if}
  {#if status.hint}<p class="muted small">{status.hint}</p>{/if}
  {#if status.error}<p class="error small">{status.error}</p>{/if}
{:else}
  <p class="muted">Loading…</p>
{/if}

<style>
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin: 0 0 var(--space-2);
  }
  .ip {
    margin-left: auto;
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--color-danger);
  }
  .dot.on {
    background: var(--color-success);
  }
  .rates {
    display: flex;
    gap: var(--space-3);
    margin: 0 0 var(--space-2);
  }
  .calls {
    list-style: none;
    margin: var(--space-2) 0 0;
    padding: 0;
  }
  .calls li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 1px 0;
  }
  .party {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .missed {
    color: var(--color-danger);
  }
  .icon {
    width: 1.2em;
    text-align: center;
  }
  .small {
    font-size: 0.8em;
  }
</style>
