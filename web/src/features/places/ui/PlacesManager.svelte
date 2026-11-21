<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { browserLocation } from '../../../shared/places/places';
  import { type Place, type PlaceCandidate, usageList, validateName } from '../domain/places';
  import type { PlacesUseCases } from '../usecase/placesUseCases';

  /** Props of the places list in Settings → Connections & places (#591). */
  interface Props {
    places: PlacesUseCases;
    /** Address search; omitted when the weather module (the geocoder) is off. */
    search?: (query: string) => Promise<PlaceCandidate[]>;
    /** Display name of a widget kind. */
    widgetName: (kind: string) => string;
    /** The account's tabs, to say where a widget is. */
    tabs: () => Promise<{ id: string; name: string }[]>;
  }
  let { places, search, widgetName, tabs }: Props = $props();

  let list = $state<Place[]>([]);
  let tabNames = $state<Map<string, string>>(new Map());
  let error = $state('');
  let query = $state('');
  let results = $state<PlaceCandidate[]>([]);
  let pending = $state<PlaceCandidate | null>(null);
  let locating = $state(false);
  let renaming = $state<{ id: string; name: string } | null>(null);
  let busy = $state(false);

  const uses = (p: Place) => usageList(p.usedBy, widgetName, (id) => tabNames.get(id));

  async function load() {
    try {
      list = await places.list();
    } catch (err) {
      error = errorMessage(err);
    }
  }
  $effect(() => {
    void load();
    tabs()
      .then((all) => (tabNames = new Map(all.map((tab) => [tab.id, tab.name]))))
      .catch(() => undefined);
  });

  async function find(e: SubmitEvent) {
    e.preventDefault();
    if (!search) return;
    error = '';
    try {
      results = await search(query);
      if (results.length === 0) error = t('places.no_results');
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function locate() {
    error = '';
    locating = true;
    try {
      const at = await browserLocation();
      pending = { name: t('places.my_location_name'), label: '', ...at };
    } catch {
      error = t('places.location_denied');
    } finally {
      locating = false;
    }
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    if (!pending) return;
    const invalid = validateName(pending.name);
    if (invalid) {
      error = t(invalid);
      return;
    }
    busy = true;
    try {
      await places.add(pending);
      pending = null;
      results = [];
      query = '';
      error = '';
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function rename(e: SubmitEvent, place: Place) {
    e.preventDefault();
    if (!renaming) return;
    const invalid = validateName(renaming.name);
    if (invalid) {
      error = t(invalid);
      return;
    }
    try {
      await places.rename(place, renaming.name);
      renaming = null;
      error = '';
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(place: Place) {
    const names = uses(place);
    const question = names.length
      ? t('places.confirm_remove_used', { name: place.name, list: names.join(', ') })
      : t('places.confirm_remove', { name: place.name });
    if (!confirm(question)) return;
    try {
      await places.remove(place.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<h3>{t('places.title')}</h3>
<p class="muted">{t('places.intro')}</p>

{#if list.length === 0}
  <p class="muted">{t('places.empty')}</p>
{:else}
  <ul class="places">
    {#each list as place (place.id)}
      <li>
        {#if renaming?.id === place.id}
          <form class="row" onsubmit={(e) => rename(e, place)}>
            <input
              class="input"
              bind:value={renaming.name}
              maxlength="80"
              aria-label={t('places.rename_place', { name: place.name })}
            />
            <button class="btn btn-primary" type="submit">{t('common.save')}</button>
            <button class="btn" type="button" onclick={() => (renaming = null)}>{t('places.cancel')}</button>
          </form>
        {:else}
          <div class="text">
            <strong>{place.name}</strong>
            <span class="muted"
              >{place.label ||
                t('places.coordinates', { lat: place.lat.toFixed(3), lon: place.lon.toFixed(3) })}</span
            >
            <span class="muted small"
              >{place.usedBy.length
                ? t('places.used_by', { list: uses(place).join(', ') })
                : t('places.unused')}</span
            >
          </div>
          <div class="actions">
            <button
              class="btn"
              type="button"
              aria-label={t('places.rename_place', { name: place.name })}
              onclick={() => (renaming = { id: place.id, name: place.name })}>{t('places.rename')}</button
            >
            <button
              class="btn btn-danger"
              type="button"
              aria-label={t('places.remove_place', { name: place.name })}
              onclick={() => remove(place)}>{t('places.remove')}</button
            >
          </div>
        {/if}
      </li>
    {/each}
  </ul>
{/if}

<h4>{t('places.add')}</h4>
{#if pending}
  <form class="row" onsubmit={save}>
    <label class="field grow">
      <span>{t('places.name')}</span>
      <input class="input" bind:value={pending.name} maxlength="80" />
    </label>
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('places.save')}</button>
    <button class="btn" type="button" onclick={() => (pending = null)}>{t('places.cancel')}</button>
  </form>
  <p class="muted small">
    {pending.label || t('places.coordinates', { lat: pending.lat.toFixed(3), lon: pending.lon.toFixed(3) })}
  </p>
{:else}
  {#if search}
    <form class="row" onsubmit={find}>
      <input
        class="input grow"
        bind:value={query}
        placeholder={t('places.search_placeholder')}
        aria-label={t('places.search_label')}
      />
      <button class="btn" type="submit">{t('places.search')}</button>
    </form>
    {#if results.length}
      <ul class="results">
        {#each results as r (`${r.lat},${r.lon}`)}
          <li>
            <button class="btn" type="button" onclick={() => ((pending = { ...r }), (results = []))}
              >{r.label || r.name}</button
            >
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
  <button
    class="btn"
    type="button"
    disabled={locating}
    aria-describedby="places-location-hint"
    onclick={locate}>{locating ? t('places.locating') : t('places.my_location')}</button
  >
  <p class="muted small" id="places-location-hint">{t('places.picker_location_hint')}</p>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .places,
  .results {
    list-style: none;
    margin: 0 0 var(--space-3);
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .places li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .small {
    font-size: 0.85rem;
  }
  .actions,
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    gap: var(--space-2);
  }
  .row {
    margin-bottom: var(--space-2);
  }
  .grow {
    flex: 1;
    min-width: 12rem;
  }
</style>
