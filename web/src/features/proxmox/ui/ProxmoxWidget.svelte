<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { type MessageKey, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import {
    availableAction,
    type ProxmoxGuest,
    type ProxmoxStatus,
    parseWidgetConfig,
    usage,
    visibleGuests,
  } from '../domain/proxmox';
  import type { ProxmoxUseCases } from '../usecase/proxmoxUseCases';

  /** Props of the Proxmox widget (#84). */
  interface Props {
    proxmox: ProxmoxUseCases;
    config: Record<string, unknown>;
  }
  let { proxmox, config }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 30_000;
  let status = $state<ProxmoxStatus | null>(null);
  let error = $state('');
  let unconfigured = $state(false);
  let busy = $state<number | null>(null);
  const statusKeys: Record<string, MessageKey> = {
    stopped: 'proxmox.status.stopped',
    paused: 'proxmox.status.paused',
    suspended: 'proxmox.status.suspended',
    prelaunch: 'proxmox.status.prelaunch',
  };
  const statusLabel = (s: string) => t(statusKeys[s] ?? 'proxmox.status.unknown');
  const guests = $derived(status ? visibleGuests(status.guests, parseWidgetConfig(config)) : []);

  const load = () =>
    proxmox
      .status()
      .then((s) => ((status = s), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });

  $effect(() => poll(load, POLL_MS));

  async function act(g: ProxmoxGuest) {
    const action = availableAction(g);
    if (!action) return;
    const question = action === 'start' ? 'proxmox.confirm_start' : 'proxmox.confirm_shutdown';
    if (!confirm(t(question, { name: g.name || String(g.id) }))) return;
    busy = g.id;
    try {
      await proxmox.act(g.id, action);
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = null;
    }
  }
</script>

{#if unconfigured}
  <p class="muted">{t('proxmox.connect_under_settings')}</p>
{:else if error && !status}
  <p class="error">{error}</p>
{:else if status}
  {#each status.nodes as n (n.name)}
    <section class="node">
      <h4>
        <span class="dot" class:off={!n.online} aria-hidden="true"></span>{n.name}
        <small class="muted"
          >{n.online
            ? t('proxmox.uptime', { value: formatUptime(n.uptimeSeconds) })
            : t('proxmox.offline')}</small
        >
      </h4>
      {#if n.online}
        <Meter label={t('proxmox.cpu', { cpus: n.cpus })} percent={n.cpuPercent} />
        <Meter
          label={t('proxmox.memory')}
          percent={usage(n.memUsed, n.memTotal)}
          detail="{formatBytes(n.memUsed)} / {formatBytes(n.memTotal)}"
        />
        <Meter
          label={t('proxmox.disk')}
          percent={usage(n.diskUsed, n.diskTotal)}
          detail="{formatBytes(n.diskUsed)} / {formatBytes(n.diskTotal)}"
        />
      {/if}
    </section>
  {/each}
  {#if guests.length}
    <ul class="guests">
      {#each guests as g (g.id)}
        {@const action = availableAction(g)}
        <li>
          <span class="dot" class:off={g.status !== 'running'} aria-hidden="true"></span>
          <span class="name" title="{g.kind === 'lxc' ? 'LXC' : 'VM'} {g.id} · {g.node}">
            <span class="muted">{g.id}</span>
            {g.name}
          </span>
          <span class="muted small"
            >{g.status === 'running' ? formatUptime(g.uptimeSeconds) : statusLabel(g.status)}</span
          >
          {#if action}
            <button
              data-action
              class="btn btn-icon"
              type="button"
              disabled={busy !== null}
              onclick={() => act(g)}
              aria-label={t(action === 'start' ? 'proxmox.start' : 'proxmox.shutdown', { name: g.name })}
              >{busy === g.id ? '…' : action === 'start' ? '▶' : '⏻'}</button
            >
          {/if}
        </li>
      {/each}
    </ul>
  {:else}
    <p class="muted">{t('proxmox.no_guests')}</p>
  {/if}
  {#if error || status.error}<p class="error small">{error || status.error}</p>{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .node h4 {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: 0 0 var(--space-1);
  }
  .node small {
    margin-left: auto;
    font-weight: normal;
  }
  .guests {
    list-style: none;
    padding: 0;
    margin: var(--space-2) 0 0;
    display: grid;
    gap: var(--space-1);
    font-size: 0.9em;
  }
  .guests li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .dot {
    width: 0.6em;
    height: 0.6em;
    border-radius: 50%;
    background: var(--color-success);
    flex: none;
  }
  .dot.off {
    background: var(--color-text-muted);
  }
  .small {
    font-size: 0.8em;
  }
</style>
