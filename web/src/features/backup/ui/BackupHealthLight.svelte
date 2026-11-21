<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { BackupStatus } from '../domain/backup';
  import { healthText } from './healthText';

  /** Props of the backup traffic light (#677). */
  interface Props {
    status: BackupStatus;
    now?: Date;
    /** Also shows the server's reason for a failure. */
    details?: boolean;
  }
  let { status, now = new Date(), details = false }: Props = $props();

  const line = $derived(healthText(status, now));
</script>

<div class="light {line.tone}" role="status">
  <span class="dot" aria-hidden="true"></span>
  <span class="text">
    <span>{line.text}</span>
    {#if details && status.health === 'failed' && status.error}
      <span class="reason">{t('backup.health.error', { error: status.error })}</span>
    {/if}
    {#if details && status.offsiteError}
      <span class="reason">{t('backup.health.offsite_error', { error: status.offsiteError })}</span>
    {/if}
  </span>
</div>

<style>
  .light {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
  }
  .dot {
    flex: none;
    width: 12px;
    height: 12px;
    margin-top: 0.3em;
    border-radius: 50%;
    background: var(--color-success);
  }
  .warn .dot {
    background: var(--color-warning);
  }
  .error .dot {
    background: var(--color-danger);
  }
  .text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .reason {
    font-size: 0.85em;
    color: var(--color-text-muted);
    overflow-wrap: anywhere;
  }
</style>
