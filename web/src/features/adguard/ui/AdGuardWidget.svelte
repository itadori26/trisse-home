<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatCount } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import {
    type AdGuardStats,
    PAUSE_MINUTES,
    type PauseMinutes,
    parseWidgetConfig,
    pauseMinutesLeft,
    providerName,
  } from '../domain/adguard';
  import type { AdGuardUseCases } from '../usecase/adguardUseCases';

  /** Props of the AdGuard widget. */
  interface Props {
    adguard: AdGuardUseCases;
    config: Record<string, unknown>;
  }
  let { adguard, config }: Props = $props();
  const provider = $derived(parseWidgetConfig(config).provider);
  const poll = usePolling();

  const POLL_MS = 60_000;
  let stats = $state<AdGuardStats | null>(null);
  let error = $state('');
  let unconfigured = $state(false);
  // Pausing (#177) is offered only when the server allows it for the provider.
  let canPause = $state(false);
  let pausing = $state(false);
  let now = $state(new Date());
  const minutesLeft = $derived(stats ? pauseMinutesLeft(stats.pausedUntil, now) : 0);

  $effect(() => {
    void adguard
      .settings(provider)
      .then((s) => (canPause = s.configured && s.canPause))
      .catch(() => (canPause = false));
  });
  // The remaining pause time counts down between polls.
  $effect(() => {
    if (!stats?.pausedUntil) return;
    const id = setInterval(() => (now = new Date()), 15_000);
    return () => clearInterval(id);
  });

  async function pause(minutes: PauseMinutes) {
    if (!confirm(t('adguard.pause_confirm', { name: providerName(provider), minutes }))) return;
    pausing = true;
    try {
      stats = await adguard.pause(provider, minutes);
      now = new Date();
      error = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      pausing = false;
    }
  }

  $effect(() => {
    const load = () =>
      adguard
        .stats(provider)
        .then((s) => ((stats = s), (now = new Date()), (error = ''), (unconfigured = false)))
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });
</script>

{#if unconfigured}
  <p class="muted">{t('adguard.connect_under_settings', { name: providerName(provider) })}</p>
{:else if error && !stats}
  <p class="error">{error}</p>
{:else if stats}
  <div class="numbers" class:two={stats.avgProcessingMs === null}>
    <div><strong>{formatCount(stats.queries)}</strong><span class="muted">{t('adguard.queries')}</span></div>
    <div><strong>{formatCount(stats.blocked)}</strong><span class="muted">{t('adguard.blocked')}</span></div>
    {#if stats.avgProcessingMs !== null}
      <div>
        <strong>{t('adguard.ms', { value: stats.avgProcessingMs.toFixed(1) })}</strong><span class="muted"
          >{t('adguard.avg')}</span
        >
      </div>
    {/if}
  </div>
  <Meter
    label={stats.protectionEnabled ? t('adguard.blocked_label') : t('adguard.protection_disabled')}
    percent={stats.blockedPercent}
    detail="{stats.blockedPercent.toFixed(1)} %"
  />
  {#if stats.topBlocked.length}
    <ol class="top">
      {#each stats.topBlocked as d (d.domain)}
        <li><span class="domain">{d.domain}</span><span class="muted">{formatCount(d.count)}</span></li>
      {/each}
    </ol>
  {/if}
  {#if minutesLeft > 0}
    <p class="paused" role="status">⏸ {t('adguard.paused_left', { minutes: minutesLeft })}</p>
  {:else if canPause && stats.protectionEnabled}
    <div class="pause" role="group" aria-label={t('adguard.pause')}>
      <span class="muted small">{t('adguard.pause')}</span>
      {#each PAUSE_MINUTES as m (m)}
        <button
          class="btn"
          type="button"
          disabled={pausing}
          aria-label={t('adguard.pause_for', { minutes: m })}
          onclick={() => pause(m)}>{t('adguard.minutes_short', { minutes: m })}</button
        >
      {/each}
    </div>
  {/if}
  {#if stats.error}<p class="error small">{stats.error}</p>{/if}
  {#if error}<p class="error small" role="alert">{error}</p>{/if}
{:else}
  <p class="muted">{t('adguard.loading')}</p>
{/if}

<style>
  .numbers {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    text-align: center;
    margin-bottom: var(--space-3);
  }
  .numbers.two {
    grid-template-columns: repeat(2, 1fr);
  }
  .numbers div {
    display: flex;
    flex-direction: column;
  }
  .numbers strong {
    font-size: clamp(1.1rem, 9cqmin, 3rem);
  }
  .numbers span,
  .small {
    font-size: 0.75em;
  }
  .pause {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    flex-wrap: wrap;
    margin: var(--space-2) 0;
  }
  .pause .btn {
    padding: 2px var(--space-2);
    font-size: 0.85em;
  }
  .paused {
    margin: var(--space-2) 0;
    color: var(--color-warning);
    font-weight: 600;
  }
  .top {
    margin: 0;
    padding-left: 1.2em;
    font-size: 0.85em;
  }
  .top li span {
    display: inline-block;
  }
  .domain {
    max-width: 75%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
    margin-right: var(--space-2);
  }
</style>
