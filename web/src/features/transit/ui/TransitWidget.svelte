<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Departure, minutesUntil, modeIcon, parseTransitConfig } from '../domain/transit';
  import type { TransitUseCases } from '../usecase/transitUseCases';

  /** Props of the transit widget. */
  interface Props {
    transit: TransitUseCases;
    config: Record<string, unknown>;
  }
  let { transit, config }: Props = $props();

  const POLL_MS = 60_000;
  const cfg = $derived(parseTransitConfig(config));
  let departures = $state<Departure[]>([]);
  let loaded = $state(false);
  let error = $state('');
  let now = $state(new Date());

  $effect(() => {
    const c = cfg;
    if (!c) return;
    const load = () => {
      // Only refresh while the page is visible: the data comes from a community service.
      if (document.visibilityState !== 'visible') return;
      transit
        .departures(c)
        .then((d) => ((departures = d), (loaded = true), (now = new Date()), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    };
    load();
    const timer = setInterval(load, POLL_MS);
    const tick = setInterval(() => (now = new Date()), 15_000);
    document.addEventListener('visibilitychange', load);
    return () => {
      clearInterval(timer);
      clearInterval(tick);
      document.removeEventListener('visibilitychange', load);
    };
  });

  const time = (d: Date) => d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
</script>

{#if !cfg}
  <p class="muted">Choose a stop in the widget settings.</p>
{:else}
  {#if cfg.stopName}<p class="stop muted">{cfg.stopName}</p>{/if}
  {#if error}<p class="error small">{error}</p>{/if}
  <ul class="deps">
    {#each departures as d, i (i)}
      <li class:cancelled={d.cancelled}>
        <span class="line" title={d.mode}>{modeIcon(d.mode)} {d.line}</span>
        <span class="dir">{d.headsign}{d.track ? ` · ${d.track}` : ''}</span>
        <span class="when" title="Planned {time(d.planned)}">
          {#if d.cancelled}
            <span class="late">cancelled</span>
          {:else}
            {minutesUntil(d, now)}
            {#if d.delayMinutes > 0}<span class="late">+{d.delayMinutes}</span>{/if}
          {/if}
        </span>
      </li>
    {:else}
      {#if loaded}<li class="muted">No departures in the near future.</li>{/if}
    {/each}
  </ul>
  <p class="source muted">Data: Transitous</p>
{/if}

<style>
  .stop {
    margin: 0 0 var(--space-1);
    font-size: 0.8em;
  }
  .deps {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    gap: var(--space-2);
    align-items: baseline;
    padding: 1px 0;
  }
  .line {
    flex: none;
    min-width: 3.6em;
    font-weight: 600;
  }
  .dir {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }
  .late {
    color: var(--color-danger);
    font-size: 0.85em;
  }
  .cancelled .dir,
  .cancelled .line {
    text-decoration: line-through;
    color: var(--color-text-muted);
  }
  .small {
    font-size: 0.8em;
  }
  .source {
    margin: var(--space-1) 0 0;
    font-size: 0.65em;
  }
</style>
