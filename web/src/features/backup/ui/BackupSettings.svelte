<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatBytes } from '../../../shared/ui/format';
  import { type BackupFile, downloadUrl } from '../domain/backup';
  import type { BackupUseCases } from '../usecase/backupUseCases';
  import OffsiteSettings from './OffsiteSettings.svelte';
  import RestoreForm from './RestoreForm.svelte';

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
  {t('backup.backups_are_created_automatically_interv')}
</p>
<p>
  <button class="btn btn-primary" type="button" disabled={busy} onclick={backupNow}
    >{busy ? t('backup.creating') : t('backup.back_up_now')}</button
  >
</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<ul class="list">
  {#each files as f (f.name)}
    <li>
      <span>{f.createdAt.toLocaleString(intlLocale())}</span>
      <span class="muted">{formatBytes(f.size)}</span>
      <a href={downloadUrl(f)} download>{t('backup.download')}</a>
    </li>
  {:else}
    <li class="muted">{t('backup.no_backups_yet')}</li>
  {/each}
</ul>

<OffsiteSettings {backups} />
<RestoreForm {backups} local={files} />

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
