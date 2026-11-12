<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { formatBytes } from '../../../shared/ui/format';
  import type { BackupFile } from '../domain/backup';
  import type { BackupUseCases } from '../usecase/backupUseCases';

  /** Props of the off-site backup target (#46). */
  interface Props {
    backups: BackupUseCases;
  }
  let { backups }: Props = $props();

  let configured = $state(false);
  let url = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let retention = $state(30);
  let remote = $state<BackupFile[]>([]);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  const input = () => ({ url, username, password, retention });

  function loadRemote() {
    backups
      .listRemote()
      .then((l) => (remote = l))
      .catch((err) => (error = errorMessage(err)));
  }

  $effect(() => {
    backups
      .remoteSettings()
      .then((s) => {
        configured = s.configured;
        url = s.url;
        username = s.username;
        hasPassword = s.hasPassword;
        retention = s.retention;
        if (s.configured) loadRemote();
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function run(action: () => Promise<void>) {
    error = message = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const save = (e: SubmitEvent) => {
    e.preventDefault();
    return run(async () => {
      const s = await backups.saveRemoteSettings(input());
      configured = s.configured;
      hasPassword = s.hasPassword;
      password = '';
      message = t('common.saved');
      loadRemote();
    });
  };
  const test = () =>
    run(async () => {
      await backups.testRemote(input());
      message = t('backup.connection_ok');
    });
  const disable = () =>
    run(async () => {
      await backups.disableRemote();
      configured = hasPassword = false;
      remote = [];
      message = t('common.saved');
    });
</script>

<h3>{t('backup.offsite')}</h3>
<p class="muted">{t('backup.offsite_hint')}</p>
<form onsubmit={save}>
  <label class="field">
    <span>{t('backup.webdav_url')}</span>
    <input
      class="input"
      type="url"
      bind:value={url}
      placeholder="https://cloud.example.org/remote.php/dav/files/USER/Backups/"
      required
    />
  </label>
  <label class="field">
    <span>{t('backup.username')}</span>
    <input class="input" bind:value={username} autocomplete="off" required />
  </label>
  <label class="field">
    <span>{t('backup.password', { value: hasPassword ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={password}
      autocomplete="new-password"
      required={!hasPassword}
    />
  </label>
  <label class="field">
    <span>{t('backup.remote_retention')}</span>
    <input class="input" type="number" min="1" max="365" bind:value={retention} />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <div class="actions">
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.test_save')}</button>
    <button class="btn" type="button" disabled={busy} onclick={test}>{t('backup.test_connection')}</button>
    {#if configured}<button class="btn" type="button" disabled={busy} onclick={disable}
        >{t('backup.disable_offsite')}</button
      >{/if}
  </div>
</form>

{#if configured && remote.length}
  <h4>{t('backup.remote_backups')}</h4>
  <ul class="list">
    {#each remote as f (f.name)}
      <li>
        <span>{f.createdAt.toLocaleString(intlLocale())}</span>
        <span class="muted">{formatBytes(f.size)}</span>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .actions {
    display: flex;
    gap: var(--space-1);
    flex-wrap: wrap;
  }
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    gap: var(--space-3);
  }
</style>
