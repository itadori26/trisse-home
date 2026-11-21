<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { useOpenSettings } from '../../../shared/ui/settingsLink';
  import type { BackupStatus } from '../domain/backup';
  import type { BackupUseCases } from '../usecase/backupUseCases';
  import BackupHealthLight from './BackupHealthLight.svelte';

  /** Props of the backup health widget (#677). */
  interface Props {
    backups: BackupUseCases;
  }
  let { backups }: Props = $props();
  const poll = usePolling();
  const openSettings = useOpenSettings();

  // Backups run hourly at most; five minutes keep the light current enough.
  const POLL_MS = 5 * 60_000;
  let status = $state<BackupStatus | null>(null);
  let now = $state(new Date());
  let error = $state('');

  $effect(() => {
    const load = () =>
      backups
        .status()
        .then((s) => ((status = s), (now = new Date()), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });
</script>

{#if error}
  <p class="error">{error}</p>
{:else if status}
  <BackupHealthLight {status} {now} details />
{/if}
{#if openSettings}
  <button type="button" class="btn open" onclick={() => openSettings('backups')}
    >{t('backup.health.open')}</button
  >
{/if}

<style>
  .open {
    margin-top: var(--space-2);
    min-height: 44px;
  }
</style>
