<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { type MaintenanceWindow, activeMaintenance } from '../domain/notifications';
  import type { NotificationsUseCases } from '../usecase/notificationsUseCases';

  /** Props of the dashboard banner for active maintenance windows. */
  interface Props {
    notifications: NotificationsUseCases;
  }
  let { notifications }: Props = $props();

  let active = $state<MaintenanceWindow[]>([]);

  $effect(() => {
    const refresh = () =>
      notifications
        .maintenance()
        .then((l) => (active = activeMaintenance(l, new Date())))
        .catch(() => (active = []));
    refresh();
    const timer = setInterval(refresh, 60_000);
    return () => clearInterval(timer);
  });
</script>

{#each active as w (w.id)}
  <p class="banner" role="status">
    🔧 {t('notifications.maintenance_banner', {
      scope: w.scope.includes('*')
        ? t('notifications.everything')
        : w.scope.map((s) => s.split(':')[1]).join(', '),
      until: w.end.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' }),
    })}{#if w.note}
      · {w.note}{/if}
  </p>
{/each}

<style>
  .banner {
    margin: 0 0 var(--space-2);
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-small);
    background: color-mix(in srgb, var(--color-warning, #b45309) 18%, transparent);
  }
</style>
