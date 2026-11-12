<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import type { MaintenanceWindow } from '../domain/notifications';
  import type { NotificationsUseCases } from '../usecase/notificationsUseCases';

  /** Props of the maintenance window section (#69). */
  interface Props {
    notifications: NotificationsUseCases;
  }
  let { notifications }: Props = $props();

  let windows = $state<MaintenanceWindow[]>([]);
  let scope = $state('*');
  let minutes = $state(60);
  let note = $state('');
  let error = $state('');

  const load = () =>
    notifications
      .maintenance()
      .then((l) => (windows = l))
      .catch((err) => (error = errorMessage(err)));

  $effect(() => {
    load();
  });

  async function start(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      await notifications.startMaintenance(scope, minutes, note);
      note = '';
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function end(id: string) {
    try {
      await notifications.endMaintenance(id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const time = (d: Date) => d.toLocaleString(intlLocale(), { dateStyle: 'short', timeStyle: 'short' });
</script>

<h3>{t('notifications.maintenance')}</h3>
<p class="muted">{t('notifications.maintenance_hint')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<ul class="list">
  {#each windows as w (w.id)}
    <li>
      <span
        ><strong>{w.scope.join(', ')}</strong> · {time(w.start)} – {time(w.end)}{#if w.note}
          · {w.note}{/if}</span
      >
      <button class="btn" type="button" onclick={() => end(w.id)}>{t('notifications.end_maintenance')}</button
      >
    </li>
  {:else}
    <li class="muted">{t('notifications.no_maintenance')}</li>
  {/each}
</ul>
<form class="start" onsubmit={start}>
  <label class="field"
    ><span>{t('notifications.maintenance_scope')}</span><input
      class="input"
      bind:value={scope}
      placeholder="service:NAS, container:nextcloud"
      required
    /></label
  >
  <label class="field"
    ><span>{t('notifications.maintenance_minutes')}</span><input
      class="input"
      type="number"
      min="1"
      max="10080"
      bind:value={minutes}
    /></label
  >
  <label class="field"
    ><span>{t('notifications.maintenance_note')}</span><input
      class="input"
      bind:value={note}
      maxlength="200"
    /></label
  >
  <button class="btn btn-primary" type="submit">{t('notifications.start_maintenance')}</button>
</form>

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
  }
  .start {
    display: grid;
    grid-template-columns: 2fr 1fr 2fr auto;
    gap: var(--space-2);
    align-items: end;
  }
  @media (max-width: 600px) {
    .start {
      grid-template-columns: 1fr;
    }
  }
</style>
