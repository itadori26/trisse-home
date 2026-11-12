<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import type { BackupFile, RestoreSource } from '../domain/backup';
  import type { BackupUseCases } from '../usecase/backupUseCases';

  /** Props of the guided restore (#46). */
  interface Props {
    backups: BackupUseCases;
    local: BackupFile[];
  }
  let { backups, local }: Props = $props();

  let remote = $state<BackupFile[]>([]);
  let choice = $state('upload');
  let file = $state<File | undefined>(undefined);
  let password = $state('');
  let code = $state('');
  let error = $state('');
  let busy = $state(false);
  let restarting = $state(false);

  $effect(() => {
    backups
      .listRemote()
      .then((l) => (remote = l))
      .catch(() => (remote = [])); // no off-site target
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!confirm(t('backup.restore_confirm'))) return;
    const [source, name] = choice.split(':', 2) as [RestoreSource, string | undefined];
    error = '';
    busy = true;
    try {
      await backups.restore({ password, code, source, name, file: source === 'upload' ? file : undefined });
      restarting = true;
      password = code = '';
      setTimeout(() => location.reload(), 8000);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const date = (f: BackupFile) => f.createdAt.toLocaleString(intlLocale());
</script>

<h3>{t('backup.restore')}</h3>
<p class="muted">{t('backup.restore_hint')}</p>
{#if restarting}
  <p role="status">{t('backup.restarting')}</p>
{:else}
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('backup.restore_source')}</span>
      <select class="input" bind:value={choice}>
        <option value="upload">{t('backup.source_upload')}</option>
        {#each local as f (f.name)}<option value={`local:${f.name}`}
            >{t('backup.source_local', { date: date(f) })}</option
          >{/each}
        {#each remote as f (f.name)}<option value={`remote:${f.name}`}
            >{t('backup.source_remote', { date: date(f) })}</option
          >{/each}
      </select>
    </label>
    {#if choice === 'upload'}
      <label class="field">
        <span>{t('backup.backup_file')}</span>
        <input
          class="input"
          type="file"
          accept=".db,application/vnd.sqlite3,application/octet-stream"
          required
          onchange={(e) => (file = (e.currentTarget as HTMLInputElement).files?.[0])}
        />
      </label>
    {/if}
    <label class="field">
      <span>{t('backup.confirm_password')}</span>
      <input class="input" type="password" bind:value={password} autocomplete="current-password" required />
    </label>
    <label class="field">
      <span>{t('backup.confirm_code')}</span>
      <input class="input" bind:value={code} inputmode="numeric" autocomplete="one-time-code" />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-danger" type="submit" disabled={busy}
      >{busy ? t('backup.restoring') : t('backup.restore_now')}</button
    >
  </form>
{/if}
