<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Place, parseWeatherConfig, placeLabel } from '../domain/weather';
  import type { WeatherUseCases } from '../usecase/weatherUseCases';

  /** Props of the weather widget settings. */
  interface Props {
    weather: WeatherUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
    /** The radar widget only needs a location. */
    mode?: 'forecast' | 'radar';
  }
  let { weather, config, onsave, mode = 'forecast' }: Props = $props();

  // svelte-ignore state_referenced_locally
  const initial = parseWeatherConfig(config);
  let selected = $state(initial ? { lat: initial.lat, lon: initial.lon } : null);
  let label = $state(initial?.label ?? '');
  let days = $state(initial?.days ?? 3);
  let showWarnings = $state(initial?.warnings ?? true);
  let query = $state('');
  let results = $state<Place[]>([]);
  let error = $state('');

  async function find(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      results = await weather.searchPlaces(query);
      if (results.length === 0) error = 'No places found.';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function choose(p: Place) {
    selected = { lat: p.lat, lon: p.lon };
    label = p.name;
    results = [];
  }

  function save() {
    if (!selected) {
      error = 'Please choose a location.';
      return;
    }
    onsave({ ...selected, label: label.trim(), days, warnings: showWarnings });
  }
</script>

<form onsubmit={find} class="search">
  <input class="input" bind:value={query} placeholder="City, e.g. Oldenburg" aria-label="Search location" />
  <button class="btn" type="submit">Search</button>
</form>
{#if results.length}
  <ul class="results">
    {#each results as p (`${p.lat},${p.lon}`)}
      <li><button class="btn" type="button" onclick={() => choose(p)}>{placeLabel(p)}</button></li>
    {/each}
  </ul>
{/if}
{#if selected}
  <p class="muted">Location: {selected.lat.toFixed(2)}, {selected.lon.toFixed(2)}</p>
{/if}
<label class="field">
  <span>Label</span>
  <input class="input" bind:value={label} maxlength="40" />
</label>
{#if mode === 'forecast'}
  <label class="field">
    <span>Forecast days: {days}</span>
    <input type="range" min="0" max="5" bind:value={days} />
  </label>
  <label
    ><input type="checkbox" bind:checked={showWarnings} /> Show official weather warnings (DWD, Germany only)</label
  >
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
<button class="btn btn-primary" type="button" onclick={save}>Save</button>

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
