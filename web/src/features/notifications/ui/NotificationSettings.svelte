<script lang="ts">
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
      message = test ? 'Saved. A test message was sent.' : 'Saved.';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if form}
  <p class="muted">
    Push messages via <a href="https://ntfy.sh" target="_blank" rel="noopener noreferrer">ntfy</a> when something
    changes for longer than two checks. Install the ntfy app and subscribe to your topic. Topics on public servers
    are public — use a long, random name or your own server with a token.
  </p>
  <form
    onsubmit={(e) => {
      e.preventDefault();
      run(false);
    }}
  >
    <label><input type="checkbox" bind:checked={form.enabled} /> Send notifications</label>
    <div class="row">
      <label class="field"
        ><span>ntfy server</span><input class="input" type="url" bind:value={form.serverUrl} /></label
      >
      <label class="field"><span>Topic</span><input class="input" bind:value={form.topic} /></label>
    </div>
    <label class="field">
      <span>Access token {hasToken ? '(leave empty to keep the stored one)' : '(optional)'}</span>
      <input class="input" type="password" autocomplete="off" bind:value={form.token} />
    </label>
    <fieldset>
      <legend>Notify me when …</legend>
      <label
        ><input type="checkbox" bind:checked={form.rules.services} /> a service goes offline or comes back</label
      >
      <label
        ><input type="checkbox" bind:checked={form.rules.containers} /> a container stops or becomes unhealthy</label
      >
      <label class="inline">
        <input type="checkbox" bind:checked={form.rules.disk} /> a disk is more than
        <input class="input num" type="number" min="50" max="99" bind:value={form.rules.diskThreshold} /> % full
      </label>
      <label class="inline">
        <input type="checkbox" bind:checked={form.rules.tls} /> a TLS certificate expires within
        <input class="input num" type="number" min="1" max="90" bind:value={form.rules.tlsDays} /> days
      </label>
      <label
        ><input type="checkbox" bind:checked={form.rules.backups} /> a monitored backup fails or is overdue</label
      >
    </fieldset>
    {#if lastError}<p class="error">Last delivery failed: {lastError}</p>{/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if message}<p class="muted">{message}</p>{/if}
    <p class="actions">
      <button class="btn btn-primary" type="submit" disabled={busy}>Save</button>
      <button class="btn" type="button" disabled={busy} onclick={() => run(true)}>Save & send test</button>
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
