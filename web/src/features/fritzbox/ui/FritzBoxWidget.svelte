<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import { type FritzBoxStatus, formatRate, internetOptions } from '../domain/fritzbox';
  import type { FritzBoxUseCases } from '../usecase/fritzboxUseCases';

  /** Props of the FRITZ!Box internet widget; the calls have their own widget (#168). */
  interface Props {
    fritzbox: FritzBoxUseCases;
    config: Record<string, unknown>;
  }
  let { fritzbox, config }: Props = $props();
  const show = $derived(internetOptions(config));
  const poll = usePolling();

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
          throw err;
        });
    return poll(load, POLL_MS);
  });
</script>

{#if unconfigured}
  <p class="muted">{t('fritzbox.connect_your_fritz_box_under_settings')}</p>
{:else if error && !status}
  <p class="error">{error}</p>
{:else if status}
  {@const wan = status.wan}
  <p class="head">
    <span class="dot" class:on={wan.connected} aria-hidden="true"></span>
    <strong>{wan.connected ? t('fritzbox.online') : t('fritzbox.offline')}</strong>
    {#if show.showUptime && wan.connected && wan.uptimeSeconds}<span class="muted small"
        >{t('fritzbox.for', { value: formatUptime(wan.uptimeSeconds) })}</span
      >{/if}
    {#if show.showIp && wan.externalIp}<span class="muted small ip" title={t('fritzbox.external_ip')}
        >{wan.externalIp}</span
      >{/if}
  </p>
  {#if show.showRates && wan.throughputKnown}
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
  {#if status.hint}<p class="muted small">{status.hint}</p>{/if}
  {#if status.error}<p class="error small">{status.error}</p>{/if}
{:else}
  <p class="muted">{t('fritzbox.loading')}</p>
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
  .small {
    font-size: 0.8em;
  }
</style>
