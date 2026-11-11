<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { suggestTopic } from '../domain/notifications';
  import type { NotificationSettingsInput, NotificationsUseCases } from '../usecase/notificationsUseCases';

  /** Props of the notification settings section. */
  interface Props {
    notifications: NotificationsUseCases;
  }
  let { notifications }: Props = $props();

  let form = $state<NotificationSettingsInput | null>(null);
  let hasToken = $state(false);
  let lastError = $state('');
  let message = $state('');
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    notifications
      .settings()
      .then((s) => {
        form = {
          enabled: s.enabled,
          serverUrl: s.serverUrl,
          topic: s.topic || suggestTopic(),
          token: '',
          rules: s.rules,
        };
        hasToken = s.hasToken;
        lastError = s.lastError;
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function run(test: boolean) {
    if (!form) return;
    busy = true;
    error = message = '';
    try {
      const s = test ? await notifications.saveAndTest(form) : await notifications.save(form);
      hasToken = s.hasToken;
      lastError = s.lastError;
      form.token = '';
      message = test ? t('notifications.saved_test_sent') : t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if form}
  <p class="muted">
    {t('notifications.push_messages_via')}
    <a href="https://ntfy.sh" target="_blank" rel="noopener noreferrer">ntfy</a>
    {t('notifications.when_something_changes_for_longer_than')}
  </p>
  <form
    onsubmit={(e) => {
      e.preventDefault();
      run(false);
    }}
  >
    <label
      ><input type="checkbox" bind:checked={form.enabled} /> {t('notifications.send_notifications')}</label
    >
    <div class="row">
      <label class="field"
        ><span>{t('notifications.ntfy_server')}</span><input
          class="input"
          type="url"
          bind:value={form.serverUrl}
        /></label
      >
      <label class="field"
        ><span>{t('notifications.topic')}</span><input class="input" bind:value={form.topic} /></label
      >
    </div>
    <label class="field">
      <span
        >{t('notifications.access_token', {
          value: hasToken ? t('common.keep_stored') : t('common.optional'),
        })}</span
      >
      <input class="input" type="password" autocomplete="off" bind:value={form.token} />
    </label>
    <fieldset>
      <legend>{t('notifications.notify_me_when')}</legend>
      <label
        ><input type="checkbox" bind:checked={form.rules.services} />
        {t('notifications.a_service_goes_offline_or_comes')}</label
      >
      <label
        ><input type="checkbox" bind:checked={form.rules.containers} />
        {t('notifications.a_container_stops_or_becomes_unhealthy')}</label
      >
      <label class="inline">
        <input type="checkbox" bind:checked={form.rules.disk} />
        {t('notifications.a_disk_is_more_than')}
        <input class="input num" type="number" min="50" max="99" bind:value={form.rules.diskThreshold} />
        {t('notifications.full')}
      </label>
      <label class="inline">
        <input type="checkbox" bind:checked={form.rules.tls} />
        {t('notifications.a_tls_certificate_expires_within')}
        <input class="input num" type="number" min="1" max="90" bind:value={form.rules.tlsDays} />
        {t('notifications.days')}
      </label>
      <label
        ><input type="checkbox" bind:checked={form.rules.backups} />
        {t('notifications.a_monitored_backup_fails_or_is')}</label
      >
      <label
        ><input type="checkbox" bind:checked={form.rules.security} />
        {t('notifications.rule_security')}</label
      >
    </fieldset>
    {#if lastError}<p class="error">
        {t('notifications.last_delivery_failed', { lastError: lastError })}
      </p>{/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if message}<p class="muted">{message}</p>{/if}
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('notifications.save')}</button>
      <button class="btn" type="button" disabled={busy} onclick={() => run(true)}
        >{t('notifications.save_send_test')}</button
      >
    </p>
  </form>
{:else if error}
  <p class="error">{error}</p>
{/if}

<style>
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    display: grid;
    gap: var(--space-1);
    margin: var(--space-3) 0;
  }
  .inline {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  .num {
    width: 4.5rem;
    padding: 2px 6px;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
  @media (max-width: 600px) {
    .row {
      grid-template-columns: 1fr;
    }
  }
</style>
