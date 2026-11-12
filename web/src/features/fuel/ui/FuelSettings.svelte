<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { fuelName, type FuelType, fuelTypes, type PlaceResult } from '../domain/fuel';
  import type { FuelUseCases } from '../usecase/fuelUseCases';

  /** Props of the fuel price settings. */
  interface Props {
    fuel: FuelUseCases;
    /** Place search of the weather feature, passed in by the app; optional. */
    searchPlaces?: (query: string) => Promise<PlaceResult[]>;
  }
  let { fuel, searchPlaces }: Props = $props();

  let apiKey = $state('');
  let hasApiKey = $state(false);
  let lat = $state(0);
  let lon = $state(0);
  let place = $state('');
  let radiusKm = $state(5);
  let fuelType = $state<FuelType>('e10');
  let query = $state('');
  let results = $state<PlaceResult[]>([]);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    fuel
      .settings()
      .then((s) => {
        hasApiKey = s.hasApiKey;
        lat = s.lat;
        lon = s.lon;
        place = s.place;
        radiusKm = s.radiusKm;
        fuelType = s.fuel;
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function search(e: SubmitEvent) {
    e.preventDefault();
    if (!searchPlaces) return;
    try {
      results = await searchPlaces(query);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  function choose(p: PlaceResult) {
    lat = p.lat;
    lon = p.lon;
    place = p.name;
    results = [];
    query = '';
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await fuel.saveSettings({ apiKey, lat, lon, place, radiusKm, fuel: fuelType });
      hasApiKey = s.hasApiKey;
      apiKey = '';
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if searchPlaces}
  <form class="search" onsubmit={search}>
    <input
      class="input"
      bind:value={query}
      placeholder={t('fuel.search_place')}
      aria-label={t('fuel.search_place')}
    />
    <button class="btn" type="submit">{t('weather.search')}</button>
  </form>
  {#if results.length}
    <ul class="results">
      {#each results as r, i (i)}
        <li>
          <button class="btn" type="button" onclick={() => choose(r)}
            >{r.name}{r.region ? `, ${r.region}` : ''}</button
          >
        </li>
      {/each}
    </ul>
  {/if}
{/if}
<form onsubmit={submit}>
  <p class="muted">
    {place
      ? t('fuel.location', { place, lat: lat.toFixed(3), lon: lon.toFixed(3) })
      : t('fuel.choose_location')}
  </p>
  {#if !searchPlaces}
    <div class="row">
      <label class="field"
        ><span>{t('fuel.lat')}</span><input class="input" type="number" step="any" bind:value={lat} /></label
      >
      <label class="field"
        ><span>{t('fuel.lon')}</span><input class="input" type="number" step="any" bind:value={lon} /></label
      >
    </div>
  {/if}
  <div class="row">
    <label class="field"
      ><span>{t('fuel.fuel')}</span><select class="input" bind:value={fuelType}>
        {#each fuelTypes as f (f)}<option value={f}>{fuelName(f)}</option>{/each}
      </select></label
    >
    <label class="field"
      ><span>{t('fuel.radius')}</span><input
        class="input"
        type="number"
        min="1"
        max="25"
        step="0.5"
        bind:value={radiusKm}
      /></label
    >
  </div>
  <label class="field">
    <span>{t('fuel.api_key', { value: hasApiKey ? t('common.keep_stored') : '' })}</span>
    <input
      class="input"
      type="password"
      bind:value={apiKey}
      autocomplete="new-password"
      required={!hasApiKey}
    />
  </label>
  <p class="muted">{t('fuel.key_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>

<style>
  .search {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .results {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }
</style>
