<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { ago, type BackupSource } from '../domain/source';
  import type { BackupMonitorUseCases } from '../usecase/backupMonitorUseCases';

  /** Props of the backup status widget. */
  interface Props {
    backupMonitor: BackupMonitorUseCases;
  }
  let { backupMonitor }: Props = $props();

  const POLL_MS = 60_000;
  let sources = $state<BackupSource[]>([]);
  let error = $state('');
  let now = $state(new Date());

  $effect(() => {
    const load = () =>
      backupMonitor
        .list()
        .then((s) => ((sources = s), (now = new Date()), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  const labels = {
    ok: 'backupmonitor.state.ok',
    overdue: 'backupmonitor.state.overdue',
    failed: 'backupmonitor.state.failed',
    waiting: 'backupmonitor.state.waiting',
  } as const;
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul class="list">
  {#each sources as s (s.id)}
    <li title={s.lastMessage}>
      <span class="dot {s.state}" aria-label={t(labels[s.state])}></span>
      <span class="name">{s.name}</span>
      <span class="muted small"
        >{s.state === 'failed'
          ? t('backupmonitor.failed_ago', { time: ago(s.lastFailureAt, now) })
          : ago(s.lastSuccessAt, now)}</span
      >
    </li>
  {:else}
    <li class="muted">{t('backupmonitor.no_backups_monitored_yet_add_them')}</li>
  {/each}
</ul>

<style>
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px 0;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
    white-space: nowrap;
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex: none;
    background: var(--color-text-muted);
  }
  .dot.ok {
    background: var(--color-success);
  }
  .dot.overdue {
    background: var(--color-warning);
  }
  .dot.failed {
    background: var(--color-danger);
  }
</style>
