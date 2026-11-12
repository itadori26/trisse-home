<script lang="ts">
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import { type Range, type SpeedResults, formatMbps, ranges } from '../domain/speedtest';
  import type { SpeedtestUseCases } from '../usecase/speedtestUseCases';

  /** Props of the speedtest widget (#60). */
  interface Props {
    speedtest: SpeedtestUseCases;
  }
  let { speedtest }: Props = $props();

  let range = $state<Range>(24);
  let data = $state<SpeedResults | null>(null);
  let error = $state('');
  let missing = $state(false);

  $effect(() => {
    const hours = range;
    const load = () =>
      speedtest
        .results(hours)
        .then((d) => ((data = d), (error = ''), (missing = false)))
        .catch((err) => {
          missing = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
        });
    load();
    const timer = setInterval(load, 10 * 60_000);
    return () => clearInterval(timer);
  });
</script>

{#if error && !data}
  <p class={missing ? 'muted' : 'error'}>{error}</p>
{:else if data}
  {#if data.latest}
    <div class="latest">
      <div>
        <span class="muted small">↓ {t('speedtest.download')}</span><strong
          >{formatMbps(data.latest.downloadMbps)}</strong
        ><span class="muted small">Mbit/s</span>
      </div>
      <div>
        <span class="muted small">↑ {t('speedtest.upload')}</span><strong
          >{formatMbps(data.latest.uploadMbps)}</strong
        ><span class="muted small">Mbit/s</span>
      </div>
      <div>
        <span class="muted small">{t('speedtest.ping')}</span><strong>{data.latest.pingMs.toFixed(0)}</strong
        ><span class="muted small">ms</span>
      </div>
    </div>
    <p class="muted small">{t('speedtest.measured', { time: data.latest.at.toLocaleString() })}</p>
  {:else}
    <p class="muted">{t('speedtest.no_results')}</p>
  {/if}
  <div class="ranges" role="group" aria-label={t('speedtest.range')}>
    {#each ranges as r (r)}
      <button class="btn" class:active={range === r} onclick={() => (range = r)} aria-pressed={range === r}
        >{t(r === 24 ? 'speedtest.range_day' : 'speedtest.range_week')}</button
      >
    {/each}
  </div>
  <Sparkline values={data.history.map((m) => m.downloadMbps)} label={t('speedtest.download_history')} />
  <Sparkline values={data.history.map((m) => m.uploadMbps)} label={t('speedtest.upload_history')} />
  <Sparkline values={data.history.map((m) => m.pingMs)} label={t('speedtest.ping_history')} />
  {#if data.error}<p class="error small">{data.error}</p>{/if}
{:else}
  <p class="muted">{t('speedtest.loading')}</p>
{/if}

<style>
  .small {
    font-size: 0.85em;
  }
  .latest {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-2);
  }
  .latest div {
    display: flex;
    flex-direction: column;
  }
  .latest strong {
    font-size: 1.4em;
    font-variant-numeric: tabular-nums;
  }
  .ranges {
    display: flex;
    gap: var(--space-1);
    margin: var(--space-1) 0;
  }
  .ranges .active {
    font-weight: 600;
  }
</style>
