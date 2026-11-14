<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { MailSettings, MailSettingsInput } from '../domain/parcel';
  import type { ParcelsUseCases } from '../usecase/parcelsUseCases';

  /** Parcels from shipping e-mails (#166): the read-only IMAP folder. */
  interface Props {
    parcels: ParcelsUseCases;
  }
  let { parcels }: Props = $props();

  let form = $state<MailSettingsInput>({
    enabled: false,
    host: '',
    port: 993,
    username: '',
    password: '',
    folder: 'INBOX',
    intervalMinutes: 15,
  });
  let status = $state<MailSettings | null>(null);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  const apply = (s: MailSettings) => {
    status = s;
    const { enabled, host, port, username, folder, intervalMinutes } = s;
    form = { enabled, host, port, username, password: '', folder, intervalMinutes };
  };
  $effect(() => {
    parcels
      .mailSettings()
      .then(apply)
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      apply(await parcels.saveMailSettings(form));
      message = form.enabled ? t('parcels.mail_saved_tested') : t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<section aria-labelledby="parcel-mail">
  <h3 id="parcel-mail">{t('parcels.mail_title')}</h3>
  <p class="muted small">
    {t('parcels.mail_hint')}
    <a
      href="https://github.com/itadori26/trisse-home/blob/main/docs/parcels-mail.md"
      target="_blank"
      rel="noopener noreferrer">{t('parcels.mail_guide')}</a
    >
  </p>
  <form onsubmit={submit}>
    <label class="check"
      ><input type="checkbox" bind:checked={form.enabled} /> {t('parcels.mail_enabled')}</label
    >
    <div class="grid">
      <label class="field">
        <span>{t('parcels.mail_host')}</span>
        <input class="input" bind:value={form.host} placeholder="imap.example.com" autocomplete="off" />
      </label>
      <label class="field">
        <span>{t('parcels.mail_port')}</span>
        <input class="input" type="number" min="1" max="65535" bind:value={form.port} />
      </label>
    </div>
    <label class="field">
      <span>{t('parcels.mail_user')}</span>
      <input class="input" bind:value={form.username} autocomplete="off" />
    </label>
    <label class="field">
      <span>{t('parcels.mail_password', { value: status?.hasPassword ? t('common.keep_stored') : '' })}</span>
      <input class="input" type="password" bind:value={form.password} autocomplete="new-password" />
    </label>
    <div class="grid">
      <label class="field">
        <span>{t('parcels.mail_folder')}</span>
        <input class="input" bind:value={form.folder} placeholder="Pakete" />
      </label>
      <label class="field">
        <span>{t('parcels.mail_interval')}</span>
        <select class="input" bind:value={form.intervalMinutes}>
          {#each [5, 15, 30, 60, 180, 720, 1440] as m (m)}
            <option value={m}
              >{t(m < 60 ? 'parcels.every_minutes' : 'parcels.every_hours', {
                n: m < 60 ? m : m / 60,
              })}</option
            >
          {/each}
        </select>
      </label>
    </div>
    {#if status?.lastRun}
      <p class="muted small" role="status">
        {t('parcels.mail_last_run', {
          time: status.lastRun.toLocaleString(intlLocale()),
          found: status.found,
        })}
      </p>
    {/if}
    {#if status?.lastError}<p class="error small">{status.lastError}</p>{/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if message}<p class="muted">{message}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('parcels.mail_save')}</button>
  </form>
</section>

<style>
  h3 {
    margin: var(--space-4) 0 var(--space-1);
    font-size: 1rem;
  }
  .grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: var(--space-2);
  }
  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .small {
    font-size: 0.85em;
  }
</style>
