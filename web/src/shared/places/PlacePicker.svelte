<script lang="ts">
  import { t } from '../i18n';
  import { browserLocation, placeAt, type SavedPlace, type SavedPlaces } from './places';

  /**
   * Props of the saved-place choice in widget settings (#591). It shows
   * nothing while the places module is off, so the form around it (search,
   * coordinates) stays the way to enter a location.
   */
  interface Props {
    places: SavedPlaces;
    /** Id of the chosen place, '' for none. */
    value: string;
    /**
     * The location the form holds now. It preselects a saved place at the
     * same spot and can be saved as a new place when it matches none.
     */
    current?: { name: string; label: string; lat: number; lon: number } | null;
    onpick: (place: SavedPlace | null) => void;
  }
  let { places, value, current = null, onpick }: Props = $props();

  let list = $state<SavedPlace[]>([]);
  // False while the places module does not answer: nothing is offered.
  let available = $state(false);
  let saving = $state(false);
  let locating = $state(false);
  let error = $state('');
  const uid = $props.id();
  const hintId = `${uid}-location-hint`;

  $effect(() => {
    places
      .list()
      .then((l) => {
        list = l;
        available = true;
        // A widget set up before places existed: preselect the place at its
        // coordinates, so it counts as using it.
        if (!value && current) {
          const match = placeAt(l, current.lat, current.lon);
          if (match) onpick(match);
        }
      })
      .catch(() => (available = false));
  });

  function change(e: Event) {
    const id = (e.currentTarget as HTMLSelectElement).value;
    onpick(list.find((p) => p.id === id) ?? null);
  }

  /**
   * Saves the browser's rounded location as a place and picks it (#693).
   * A place already saved at that spot is reused instead of duplicated.
   */
  async function useMyLocation() {
    locating = true;
    error = '';
    try {
      let at: { lat: number; lon: number };
      try {
        at = await browserLocation();
      } catch {
        error = t('places.picker_location_denied');
        return;
      }
      const existing = placeAt(list, at.lat, at.lon);
      const p =
        existing ?? (await places.add({ name: t('places.picker_my_location_name'), label: '', ...at }));
      if (!existing) list = [...list, p];
      onpick(p);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      locating = false;
    }
  }

  async function save() {
    if (!current) return;
    saving = true;
    error = '';
    try {
      const p = await places.add({
        name: current.name || current.label,
        label: current.label,
        lat: current.lat,
        lon: current.lon,
      });
      list = [...list, p];
      onpick(p);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      saving = false;
    }
  }
</script>

{#if available}
  {#if list.length}
    <label class="field">
      <span>{t('places.saved_place')}</span>
      <select class="input" {value} onchange={change}>
        <option value="">{t('places.other_location')}</option>
        {#each list as p (p.id)}
          <option value={p.id}>{p.label ? `${p.name} – ${p.label}` : p.name}</option>
        {/each}
      </select>
    </label>
  {/if}
  <div class="actions">
    <button class="btn" type="button" disabled={locating} aria-describedby={hintId} onclick={useMyLocation}
      >{locating ? t('places.picker_locating') : t('places.picker_my_location')}</button
    >
    {#if !value && current && (current.name || current.label)}
      <button class="btn" type="button" disabled={saving} onclick={save}>{t('places.save_as_place')}</button>
    {/if}
  </div>
  <p class="muted hint" id={hintId}>{t('places.picker_location_hint')}</p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
{/if}

<style>
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .hint {
    font-size: 0.85rem;
    margin: var(--space-1) 0 var(--space-2);
  }
</style>
