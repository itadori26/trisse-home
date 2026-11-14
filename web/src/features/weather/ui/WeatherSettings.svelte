<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    type HealthOptions,
    type Place,
    parseWeatherConfig,
    placeLabel,
    POLLEN_TYPES,
    type PollenType,
  } from '../domain/weather';
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
  // Pollen and UV index (#56): off until a region or city is chosen.
  let showHealth = $state(Boolean(initial && (initial.pollenRegion > 0 || initial.uvCity !== '')));
  let pollenRegion = $state(initial?.pollenRegion ?? 0);
  let uvCity = $state(initial?.uvCity ?? '');
  let pollenTypes = $state<PollenType[]>(initial?.pollenTypes ?? []);
  let options = $state<HealthOptions | null>(null);
  let optionsError = $state('');

  // The DWD lists are only loaded once the section is switched on.
  $effect(() => {
    if (!showHealth || options) return;
    weather
      .healthOptions()
      .then((o) => (options = o))
      .catch((err) => (optionsError = errorMessage(err)));
  });

  function togglePollen(p: PollenType, on: boolean) {
    pollenTypes = on
      ? POLLEN_TYPES.filter((x) => x === p || pollenTypes.includes(x))
      : pollenTypes.filter((x) => x !== p);
  }
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
    const health = showHealth
      ? { pollenRegion, uvCity, pollenTypes }
      : { pollenRegion: 0, uvCity: '', pollenTypes: [] };
    onsave({ ...selected, label: label.trim(), days, warnings: showWarnings, ...health });
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
  <label><input type="checkbox" bind:checked={showHealth} /> {t('weather.show_pollen_uv')}</label>
  {#if showHealth}
    <fieldset class="health">
      <legend>{t('weather.pollen_uv_dwd')}</legend>
      {#if optionsError}<p class="error">{optionsError}</p>{/if}
      <label class="field"
        ><span>{t('weather.pollen_region')}</span><select class="input" bind:value={pollenRegion}>
          <option value={0}>{t('weather.none')}</option>
          {#each options?.regions ?? [] as r (r.id)}<option value={r.id}>{r.name}</option>{/each}
        </select></label
      >
      {#if pollenRegion > 0}
        <div class="types" role="group" aria-label={t('weather.pollen_types')}>
          <span class="muted small">{t('weather.pollen_types_hint')}</span>
          {#each POLLEN_TYPES as p (p)}
            <label
              ><input
                type="checkbox"
                checked={pollenTypes.includes(p)}
                onchange={(e) => togglePollen(p, e.currentTarget.checked)}
              />
              {t(`weather.pollen_${p}`)}</label
            >
          {/each}
        </div>
      {/if}
      <label class="field"
        ><span>{t('weather.uv_city')}</span><select class="input" bind:value={uvCity}>
          <option value="">{t('weather.none')}</option>
          {#each options?.cities ?? [] as c (c)}<option value={c}>{c}</option>{/each}
        </select></label
      >
    </fieldset>
  {/if}
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
<button class="btn btn-primary" type="button" onclick={save}>{t('weather.save')}</button>

<style>
  .health {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    display: grid;
    gap: var(--space-1);
    margin: var(--space-2) 0;
  }
  .types {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-3);
  }
  .types .small {
    flex-basis: 100%;
    font-size: 0.85em;
  }
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
