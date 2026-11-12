<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type BackupSource, type CreatedSource, type SourceSchedule, pingExamples } from '../domain/source';
  import type { BackupMonitorUseCases } from '../usecase/backupMonitorUseCases';

  /** Props of the backup source management section. */
  interface Props {
    backupMonitor: BackupMonitorUseCases;
  }
  let { backupMonitor }: Props = $props();

  let sources = $state<BackupSource[]>([]);
  let created = $state<CreatedSource | null>(null);
  let name = $state('');
  let schedule = $state<SourceSchedule>({ kind: 'backup', maxAgeHours: 26, cron: '', graceMinutes: 30 });
  let useCron = $state(false);
  let error = $state('');

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      sources = await backupMonitor.list();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    run(async () => {});
  });
</script>

<p class="muted">
  {t('backupmonitor.each_source_gets_a_secret_ping')}
</p>
{#if created}
  <div class="created card">
    <p>
      <strong>{t('backupmonitor.ping_url_for', { name: created.source.name })}</strong>
      {t('backupmonitor.shown_only_now')}
    </p>
    <pre>{pingExamples(created.pingUrl)}</pre>
    <button class="btn" type="button" onclick={() => (created = null)}>{t('backupmonitor.done')}</button>
  </div>
{/if}
<ul class="list">
  {#each sources as s (s.id)}
    <li>
      <span class="name"
        >{s.name}
        <span class="muted"
          >{s.kind === 'check' ? t('backupmonitor.kind_check') : t('backupmonitor.kind_backup')} ·
          {s.cron ? s.cron : t('backupmonitor.every_h', { hours: s.maxAgeHours })}{#if s.graceMinutes}
            · {t('backupmonitor.grace_short', { minutes: s.graceMinutes })}{/if}
          · {t(`backupmonitor.state.${s.state}`)}</span
        ></span
      >
      <button
        class="btn btn-icon"
        type="button"
        onclick={() =>
          confirm(t('backupmonitor.confirm_rotate', { name: s.name })) &&
          run(async () => (created = await backupMonitor.rotateToken(s.id)))}
        >{t('backupmonitor.new_url')}</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('backupmonitor.delete', { name: s.name })}
        onclick={() =>
          confirm(t('common.confirm_delete', { name: s.name })) && run(() => backupMonitor.delete(s.id))}
        >✕</button
      >
    </li>
  {/each}
</ul>
<form
  class="add"
  onsubmit={(e) => {
    e.preventDefault();
    run(async () => {
      created = await backupMonitor.create(name, { ...schedule, cron: useCron ? schedule.cron : '' });
      name = '';
    });
  }}
>
  <input
    class="input"
    bind:value={name}
    placeholder={t('backupmonitor.e_g_nas_restic')}
    aria-label={t('backupmonitor.name')}
    required
    maxlength="60"
  />
  <select class="input" bind:value={schedule.kind} aria-label={t('backupmonitor.kind')}>
    <option value="backup">{t('backupmonitor.kind_backup')}</option>
    <option value="check">{t('backupmonitor.kind_check')}</option>
  </select>
  <select class="input" bind:value={useCron} aria-label={t('backupmonitor.schedule')}>
    <option value={false}>{t('backupmonitor.schedule_interval')}</option>
    <option value={true}>{t('backupmonitor.schedule_cron')}</option>
  </select>
  {#if useCron}
    <input
      class="input cron"
      bind:value={schedule.cron}
      placeholder="0 3 * * *"
      aria-label={t('backupmonitor.cron')}
      required
    />
  {:else}
    <label class="hours"
      ><input
        class="input"
        type="number"
        min="1"
        max="1440"
        bind:value={schedule.maxAgeHours}
        aria-label={t('backupmonitor.maximum_age_in_hours')}
      /> h</label
    >
  {/if}
  <label class="hours"
    >+<input
      class="input"
      type="number"
      min="0"
      max="10080"
      bind:value={schedule.graceMinutes}
      aria-label={t('backupmonitor.grace_minutes')}
    /> min</label
  >
  <button class="btn" type="submit">{t('backupmonitor.add')}</button>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    align-items: center;
  }
  .name {
    flex: 1;
  }
  .add {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
  }
  .cron {
    width: 9rem;
    font-family: monospace;
  }
  .hours {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .hours .input {
    width: 5rem;
  }
  .created {
    padding: var(--space-3);
    margin-bottom: var(--space-3);
  }
  pre {
    overflow: auto;
    font-size: 0.8rem;
    background: var(--color-surface-2);
    padding: var(--space-2);
    border-radius: var(--radius-small);
  }
</style>
