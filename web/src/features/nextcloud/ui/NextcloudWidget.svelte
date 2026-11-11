<script lang="ts">
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatCount } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import { type NextcloudStatus, shortVersion } from '../domain/nextcloud';
  import type { NextcloudUseCases } from '../usecase/nextcloudUseCases';

  /** Props of the Nextcloud widget. */
  interface Props {
    nextcloud: NextcloudUseCases;
  }
  let { nextcloud }: Props = $props();

  const POLL_MS = 5 * 60_000;
  let status = $state<NextcloudStatus | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  $effect(() => {
    const load = () =>
      nextcloud
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
</script>

{#if unconfigured}
  <p class="muted">Connect Nextcloud under Settings → Nextcloud.</p>
{:else if error && !status}
  <p class="error">{error}</p>
{:else if status}
  <p class="head">
    <strong>Nextcloud {shortVersion(status.version)}</strong>
    {#if status.maintenance}<span class="badge warn">maintenance</span>{/if}
    {#if status.server?.appUpdates}<span class="badge">{status.server.appUpdates} app updates</span>{/if}
  </p>
  {#if status.quota.total > 0}
    <Meter
      label="Your quota"
      percent={status.quota.percent}
      detail="{formatBytes(status.quota.used)} of {formatBytes(status.quota.total)}"
    />
  {:else}
    <p class="muted small">You use {formatBytes(status.quota.used)} (no quota).</p>
  {/if}
  {#if status.server}
    <div class="numbers">
      <div><strong>{formatBytes(status.server.freeSpace)}</strong><span class="muted">free</span></div>
      <div>
        <strong>{status.server.activeUsers1d}/{status.server.users}</strong><span class="muted"
          >active today</span
        >
      </div>
      <div><strong>{formatCount(status.server.files)}</strong><span class="muted">files</span></div>
    </div>
  {:else if status.serverHint}
    <p class="muted small">{status.serverHint}</p>
  {/if}
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
  .badge {
    font-size: 0.75em;
    padding: 1px 6px;
    border-radius: 999px;
    background: var(--color-surface-2);
  }
  .badge.warn {
    color: var(--color-warning);
  }
  .numbers {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-2);
    margin-top: var(--space-2);
    text-align: center;
  }
  .numbers div {
    display: flex;
    flex-direction: column;
  }
  .numbers span,
  .small {
    font-size: 0.8em;
  }
</style>
