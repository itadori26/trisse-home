<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import {
    canWake,
    certificateDaysLeft,
    type LatencyPoint,
    type Service,
    statusText,
  } from '../domain/service';
  import type { ServicesUseCases } from '../usecase/servicesUseCases';
  import ServiceIcon from './ServiceIcon.svelte';

  /** Props of the service tiles widget. */
  interface Props {
    services: ServicesUseCases;
    config: Record<string, unknown>;
  }
  let { services, config }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 30_000;
  let list = $state<Service[]>([]);
  let history = $state<Record<string, LatencyPoint[]>>({});
  let notice = $state('');

  async function wake(s: Service) {
    if (!confirm(t('services.confirm_wake', { name: s.name }))) return;
    try {
      await services.wake(s.id);
      notice = t('services.wake_sent', { name: s.name });
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }
  let error = $state('');
  const compact = $derived(config['compact'] === true);

  $effect(() => {
    const load = () => {
      services
        .history()
        .then((h) => (history = h))
        .catch(() => {});
      return services
        .list()
        .then((l) => ((list = l), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    };
    return poll(load, POLL_MS);
  });
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if notice}<p class="muted small" role="status">{notice}</p>{/if}
{#if list.length === 0 && !error}
  <p class="muted">{t('services.no_services_yet_add_them_under')}</p>
{/if}
<ul class="tiles" class:compact>
  {#each list as s (s.id)}
    <li>
      <a href={s.url} target="_blank" rel="noopener noreferrer" title={statusText(s)}>
        <ServiceIcon icon={s.icon} name={s.name} />
        <span class="text">
          <span class="name">{s.name}</span>
          {#if s.description && !compact}<span class="muted desc">{s.description}</span>{/if}
          {#if history[s.id]?.length}
            <Sparkline
              values={(history[s.id] ?? []).map((p) => p.latencyMs)}
              label={t('services.response_time_of_over_24_hours', { name: s.name })}
            />
          {/if}
        </span>
        {#if s.certExpiresSoon}
          {@const days = certificateDaysLeft(s, new Date())}
          <span
            class="cert"
            title={t('services.tls_certificate_expires', {
              value: s.certExpiresAt?.toLocaleDateString(intlLocale()) ?? '',
            })}>🔒 {days}d</span
          >
        {/if}
        <span class="dot {s.state}" aria-label={s.state}></span>
      </a>
      {#if canWake(s)}
        <button
          class="btn btn-icon wake"
          type="button"
          title={t('services.wake_up_wake_on_lan')}
          aria-label={t('services.wake_up', { name: s.name })}
          onclick={() => wake(s)}>⏻</button
        >
      {/if}
    </li>
  {/each}
</ul>

<style>
  li {
    position: relative;
  }
  .wake {
    position: absolute;
    right: 4px;
    bottom: 2px;
    padding: 0 4px;
    font-size: 0.85em;
    line-height: 1.2;
  }
  .small {
    font-size: 0.8em;
  }
  .tiles {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: var(--space-2);
  }
  .tiles.compact {
    grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  }
  a {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2);
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    color: inherit;
    text-decoration: none;
  }
  a:hover {
    outline: 1px solid var(--color-accent);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }
  .name,
  .desc {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .desc {
    font-size: 0.8em;
  }
  .cert {
    font-size: 0.75em;
    color: var(--color-warning);
    white-space: nowrap;
  }
  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex: none;
    background: var(--color-text-muted);
  }
  .dot.online {
    background: var(--color-success);
  }
  .dot.offline {
    background: var(--color-danger);
  }
</style>
