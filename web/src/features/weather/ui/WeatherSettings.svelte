<script lang="ts">
  import { t } from '../../../shared/i18n';
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
      if (results.length === 0) error = t('weather.no_places_found');
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
      error = t('weather.choose_location');
      return;
    }
    onsave({ ...selected, label: label.trim(), days, warnings: showWarnings });
  }
</script>

<form onsubmit={find} class="search">
  <input
    class="input"
    bind:value={query}
    placeholder={t('weather.city_e_g_oldenburg')}
    aria-label={t('weather.search_location')}
  />
  <button class="btn" type="submit">{t('weather.search')}</button>
</form>
{#if results.length}
  <ul class="results">
    {#each results as p (`${p.lat},${p.lon}`)}
      <li><button class="btn" type="button" onclick={() => choose(p)}>{placeLabel(p)}</button></li>
    {/each}
  </ul>
{/if}
{#if selected}
  <p class="muted">
    {t('weather.location', { value: selected.lat.toFixed(2), value2: selected.lon.toFixed(2) })}
  </p>
{/if}
<label class="field">
  <span>{t('weather.label')}</span>
  <input class="input" bind:value={label} maxlength="40" />
</label>
{#if mode === 'forecast'}
  <label class="field">
    <span>{t('weather.forecast_days', { days: days })}</span>
    <input type="range" min="0" max="5" bind:value={days} />
  </label>
  <label
    ><input type="checkbox" bind:checked={showWarnings} />
    {t('weather.show_official_weather_warnings_dwd_germa')}</label
  >
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
<button class="btn btn-primary" type="button" onclick={save}>{t('weather.save')}</button>

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
