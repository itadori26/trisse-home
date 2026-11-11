<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatBytes } from '../../../shared/ui/format';
  import { type BackupFile, downloadUrl } from '../domain/backup';
  import type { BackupUseCases } from '../usecase/backupUseCases';

  /** Props of the backup settings section. */
  interface Props {
    backups: BackupUseCases;
  }
  let { backups }: Props = $props();

  let files = $state<BackupFile[]>([]);
  let error = $state('');
  let busy = $state(false);

  async function load() {
    try {
      files = await backups.list();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function backupNow() {
    busy = true;
    error = '';
    try {
      await backups.backupNow();
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  $effect(() => {
    load();
  });
</script>

<p class="muted">
  Backups are created automatically (interval and retention are configured with BACKUP_INTERVAL and
  BACKUP_RETENTION). Stored passwords are encrypted with SECRET_KEY — keep that key safe separately.
</p>
<p>
  <button class="btn btn-primary" type="button" disabled={busy} onclick={backupNow}
    >{busy ? 'Creating…' : 'Back up now'}</button
  >
</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<ul class="list">
  {#each files as f (f.name)}
    <li>
      <span>{f.createdAt.toLocaleString()}</span>
      <span class="muted">{formatBytes(f.size)}</span>
      <a href={downloadUrl(f)} download>Download</a>
    </li>
  {:else}
    <li class="muted">No backups yet.</li>
  {/each}
</ul>

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    gap: var(--space-3);
    align-items: center;
    padding: 2px 0;
  }
  .list li span:first-child {
    flex: 1;
  }
</style>
