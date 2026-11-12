<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import type { Service } from '../domain/service';
  import {
    type Incident,
    type ServiceUptime,
    formatDuration,
    formatUptime,
    uptimeLevel,
  } from '../domain/uptime';
  import type { ServicesUseCases } from '../usecase/servicesUseCases';

  /** Props of the status overview widget (#81). */
  interface Props {
    services: ServicesUseCases;
  }
  let { services }: Props = $props();
  const poll = usePolling();

  let list = $state<Service[]>([]);
  let uptime = $state<Record<string, ServiceUptime>>({});
  let open = $state<string | null>(null);
  let incidents = $state<Incident[]>([]);
  let error = $state('');

  const load = () =>
    Promise.all([services.list(), services.uptime()])
      .then(([l, u]) => {
        list = l;
        uptime = Object.fromEntries(u.map((x) => [x.serviceId, x]));
        error = '';
      })
      .catch((err) => {
        error = errorMessage(err);
        throw err;
      });

  $effect(() => {
    return poll(load, 5 * 60_000);
  });

  async function toggle(id: string) {
    if (open === id) {
      open = null;
      return;
    }
    open = id;
    incidents = [];
    try {
      incidents = await services.incidents(id, 10);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const when = (d: Date) => d.toLocaleString(intlLocale(), { dateStyle: 'short', timeStyle: 'short' });
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul class="status">
  {#each list as s (s.id)}
    {@const u = uptime[s.id]}
    <li>
      <button class="row" type="button" aria-expanded={open === s.id} onclick={() => toggle(s.id)}>
        <span class="dot {s.state}"></span>
        <span class="name">{s.name}</span>
        <span class="pct" title={t('services.uptime_30d')}>{formatUptime(u?.month ?? null)}</span>
      </button>
      {#if u}
        <div class="bars" role="img" aria-label={t('services.uptime_bars', { name: s.name })}>
          {#each u.days.slice(-30) as d, i (i)}
            <span class="bar {uptimeLevel(d)}" title={formatUptime(d)}></span>
          {/each}
        </div>
      {/if}
      {#if open === s.id && u}
        <dl class="windows">
          <dt>24 h</dt>
          <dd>{formatUptime(u.day)}</dd>
          <dt>7 d</dt>
          <dd>{formatUptime(u.week)}</dd>
          <dt>30 d</dt>
          <dd>{formatUptime(u.month)}</dd>
          <dt>90 d</dt>
          <dd>{formatUptime(u.quarter)}</dd>
        </dl>
        <ul class="incidents">
          {#each incidents as i (i.id)}
            <li>
              <span
                >{when(i.start)} · {i.end
                  ? formatDuration(i.durationSeconds)
                  : t('services.incident_ongoing')}</span
              >
              {#if i.cause}<span class="muted small">{i.cause}</span>{/if}
            </li>
          {:else}
            <li class="muted">{t('services.no_incidents')}</li>
          {/each}
        </ul>
      {/if}
    </li>
  {:else}
    <li class="muted">{t('services.no_services_yet_add_them_under')}</li>
  {/each}
</ul>

<style>
  .status {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .status > li {
    padding: var(--space-1) 0;
  }
  .row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    background: none;
    border: 0;
    color: inherit;
    font: inherit;
    padding: 0;
    cursor: pointer;
    text-align: left;
  }
  .name {
    flex: 1;
  }
  .pct {
    font-variant-numeric: tabular-nums;
  }
  .dot {
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    background: var(--color-muted, #888);
  }
  .dot.online {
    background: var(--color-success);
  }
  .dot.offline {
    background: var(--color-danger);
  }
  .bars {
    display: flex;
    gap: 2px;
    height: 0.9rem;
    margin-top: 2px;
  }
  .bar {
    flex: 1;
    border-radius: 1px;
    background: var(--color-border, rgba(127, 127, 127, 0.25));
  }
  .bar.good {
    background: var(--color-success);
  }
  .bar.degraded {
    background: var(--color-warning);
  }
  .bar.bad {
    background: var(--color-danger);
  }
  .windows {
    display: grid;
    grid-template-columns: repeat(4, auto auto);
    gap: 0 var(--space-2);
    margin: var(--space-1) 0;
    font-size: 0.85em;
  }
  .windows dd {
    margin: 0;
    font-variant-numeric: tabular-nums;
  }
  .incidents {
    list-style: none;
    padding: 0;
    font-size: 0.85em;
  }
  .incidents li {
    display: flex;
    flex-direction: column;
  }
  .small {
    font-size: 0.9em;
  }
</style>
