<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Parcel, parcelName, statusInfo } from '../domain/parcel';
  import type { ParcelsUseCases } from '../usecase/parcelsUseCases';

  /** Props of the parcels widget. */
  interface Props {
    parcels: ParcelsUseCases;
  }
  let { parcels }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 10 * 60_000;
  let list = $state<Parcel[]>([]);
  let loaded = $state(false);
  let adding = $state(false);
  let number = $state('');
  let label = $state('');
  let error = $state('');

  const load = () =>
    parcels
      .list()
      .then((l) => ((list = l), (loaded = true), (error = '')))
      .catch((err) => {
        error = errorMessage(err);
        throw err;
      });

  $effect(() => {
    return poll(load, POLL_MS);
  });

  async function add(e: SubmitEvent) {
    e.preventDefault();
    try {
      await parcels.add(number, label);
      number = label = '';
      adding = false;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(p: Parcel) {
    if (!confirm(t('parcels.confirm_stop', { name: parcelName(p) }))) return;
    try {
      await parcels.remove(p.id);
      list = list.filter((x) => x.id !== p.id);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const day = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'numeric' });
</script>

<div class="head">
  <button class="btn btn-icon" type="button" aria-expanded={adding} onclick={() => (adding = !adding)}
    >{adding ? '✕' : t('parcels.add_button')}</button
  >
</div>
{#if adding}
  <form class="add" onsubmit={add}>
    <input
      class="input"
      bind:value={number}
      placeholder={t('parcels.tracking_number_dhl')}
      aria-label={t('parcels.tracking_number')}
      required
    />
    <input
      class="input"
      bind:value={label}
      placeholder={t('parcels.label_optional')}
      aria-label={t('parcels.label')}
      maxlength="60"
    />
    <button class="btn btn-primary" type="submit">{t('parcels.track')}</button>
  </form>
{/if}
{#if error}<p class="error small" role="alert">{error}</p>{/if}
<ul class="parcels">
  {#each list as p (p.id)}
    {@const info = statusInfo[p.status]}
    <li>
      <div class="row">
        <span class="name" title={p.trackingNumber}>{parcelName(p)}</span>
        <span class="badge" style="color: {info.color}">{t(info.label)}</span>
        <button
          class="btn btn-icon rm"
          type="button"
          aria-label={t('parcels.remove', { value: parcelName(p) })}
          onclick={() => remove(p)}>✕</button
        >
      </div>
      {#if p.error}
        <div class="muted small">{p.error}</div>
      {:else}
        <div class="muted small">
          {p.description}{p.location ? ` · ${p.location}` : ''}
          {#if p.estimatedDelivery && p.status !== 'delivered'}
            · expected {day(p.estimatedDelivery)}{/if}
        </div>
      {/if}
    </li>
  {:else}
    {#if loaded}<li class="muted">{t('parcels.no_parcels_tracked')}</li>{/if}
  {/each}
</ul>

<style>
  .head {
    display: flex;
    justify-content: flex-end;
  }
  .add {
    display: grid;
    gap: var(--space-1);
    margin: var(--space-1) 0 var(--space-2);
  }
  .parcels {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: 600;
  }
  .badge {
    font-size: 0.8em;
    white-space: nowrap;
  }
  .rm {
    padding: 0 4px;
  }
  .small {
    font-size: 0.8em;
  }
</style>
