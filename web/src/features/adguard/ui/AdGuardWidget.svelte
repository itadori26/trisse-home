<script lang="ts">
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatCount } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import type { AdGuardStats } from '../domain/adguard';
  import type { AdGuardUseCases } from '../usecase/adguardUseCases';

  /** Props of the AdGuard widget. */
  interface Props {
    adguard: AdGuardUseCases;
  }
  let { adguard }: Props = $props();

  const POLL_MS = 60_000;
  let stats = $state<AdGuardStats | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  $effect(() => {
    const load = () =>
      adguard
        .stats()
        .then((s) => ((stats = s), (error = ''), (unconfigured = false)))
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
        });
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });
</script>

{#if unconfigured}
  <p class="muted">Connect AdGuard Home under Settings → AdGuard.</p>
{:else if error && !stats}
  <p class="error">{error}</p>
{:else if stats}
  <div class="numbers">
    <div><strong>{formatCount(stats.queries)}</strong><span class="muted">queries</span></div>
    <div><strong>{formatCount(stats.blocked)}</strong><span class="muted">blocked</span></div>
    <div><strong>{stats.avgProcessingMs.toFixed(1)} ms</strong><span class="muted">avg</span></div>
  </div>
  <Meter
    label={stats.protectionEnabled ? 'Blocked' : 'Protection disabled!'}
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
  {#if stats.error}<p class="error small">{stats.error}</p>{/if}
{:else}
  <p class="muted">Loading…</p>
{/if}

<style>
  .numbers {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    text-align: center;
    margin-bottom: var(--space-3);
  }
  .numbers div {
    display: flex;
    flex-direction: column;
  }
  .numbers strong {
    font-size: 1.2rem;
  }
  .numbers span,
  .small {
    font-size: 0.75rem;
  }
  .top {
    margin: 0;
    padding-left: 1.2em;
    font-size: 0.85rem;
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
