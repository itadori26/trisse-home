<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import {
    type Favourite,
    formatPrice,
    fuelName,
    type FuelOverview,
    isFavourite,
    parseWidgetConfig,
    type Station,
    toggleFavourite,
  } from '../domain/fuel';
  import type { FuelUseCases } from '../usecase/fuelUseCases';

  /** Props of the fuel price widget (#53). */
  interface Props {
    fuel: FuelUseCases;
    config: Record<string, unknown>;
  }
  let { fuel, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseWidgetConfig(config));

  const POLL_MS = 5 * 60_000;
  let overview = $state<FuelOverview | null>(null);
  let favourites = $state<Favourite[]>([]);
  let error = $state('');
  let unconfigured = $state(false);
  const list = $derived(
    overview ? (cfg.view === 'favourites' ? overview.favourites : overview.cheapest) : [],
  );

  const load = () =>
    Promise.all([fuel.overview(), fuel.settings()])
      .then(([o, s]) => ((overview = o), (favourites = s.favourites), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });

  $effect(() => poll(load, POLL_MS));

  async function star(s: Station) {
    try {
      favourites = (await fuel.setFavourites(toggleFavourite(favourites, s))).favourites;
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if unconfigured}
  <p class="muted">{t('fuel.connect_under_settings')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  <p class="muted small">{fuelName(overview.fuel)}{overview.place ? ` · ${overview.place}` : ''}</p>
  <ul class="stations">
    {#each list as s (s.id)}
      {@const p = formatPrice(s.price, intlLocale())}
      <li class:closed={!s.open}>
        <span class="price"
          >{#if p}{p.main}<sup>{p.tenth}</sup>{:else}–{/if}</span
        >
        <span class="name" title={s.address}>
          {s.brand || s.name}
          <small class="muted"
            >{s.open ? '' : `${t('fuel.closed')} · `}{s.distanceKm > 0
              ? `${s.distanceKm} km`
              : s.address}</small
          >
        </span>
        <button
          class="btn btn-icon"
          type="button"
          aria-pressed={isFavourite(favourites, s.id)}
          aria-label={t('fuel.favourite', { name: s.brand || s.name })}
          onclick={() => star(s)}>{isFavourite(favourites, s.id) ? '★' : '☆'}</button
        >
      </li>
    {:else}
      <li class="muted">{cfg.view === 'favourites' ? t('fuel.no_favourites') : t('fuel.no_stations')}</li>
    {/each}
  </ul>
  {#if error}<p class="error small">{error}</p>{/if}
  <p class="muted tiny">{t('fuel.attribution')}</p>
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .stations {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    gap: var(--space-1);
  }
  .stations li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .stations li.closed {
    opacity: 0.6;
  }
  .price {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    min-width: 3.6em;
  }
  .name {
    flex: 1;
    min-width: 0;
    display: grid;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
    margin: 0 0 var(--space-1);
  }
  .tiny {
    font-size: 0.7em;
    margin: var(--space-1) 0 0;
  }
</style>
