<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { parseTransitConfig, type Stop } from '../domain/transit';
  import type { TransitUseCases } from '../usecase/transitUseCases';

  /** Props of the transit widget settings. */
  interface Props {
    transit: TransitUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { transit, config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  const initial = parseTransitConfig(config);
  let stop = $state(initial ? { id: initial.stopId, name: initial.stopName } : null);
  let count = $state(initial?.count ?? 6);
  let query = $state('');
  let results = $state<Stop[]>([]);
  let error = $state('');

  async function find(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      results = await transit.searchStops(query);
      if (results.length === 0) error = t('transit.no_stops_found');
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function save() {
    if (!stop) {
      error = t('transit.choose_stop');
      return;
    }
    onsave({ stopId: stop.id, stopName: stop.name, count });
  }
</script>

<form onsubmit={find} class="search">
  <input
    class="input"
    bind:value={query}
    placeholder={t('transit.stop_e_g_oldenburg_hbf')}
    aria-label={t('transit.search_stop')}
  />
  <button class="btn" type="submit">{t('transit.search')}</button>
</form>
{#if results.length}
  <ul class="results">
    {#each results as s (s.id)}
      <li>
        <button
          class="btn"
          type="button"
          onclick={() => ((stop = { id: s.id, name: s.name }), (results = []))}>{s.name}</button
        >
      </li>
    {/each}
  </ul>
{/if}
{#if stop}<p class="muted">{t('transit.stop', { name: stop.name })}</p>{/if}
<label class="field">
  <span>{t('transit.departures', { count: count })}</span>
  <input type="range" min="1" max="20" bind:value={count} />
</label>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<button class="btn btn-primary" type="button" onclick={save}>{t('transit.save')}</button>

<style>
  .search {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }
  .results {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .results .btn {
    width: 100%;
  }
</style>
