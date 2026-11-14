<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { CALL_FILTERS, callsOptions } from '../domain/fritzbox';

  /** Settings of the FRITZ!Box calls widget (#168). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  const initial = callsOptions(config);
  let filter = $state(initial.filter);
  let limit = $state(initial.limit);
  let showNumbers = $state(initial.showNumbers);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...config, filter, limit, showNumbers });
  }}
>
  <label class="field">
    <span>{t('fritzbox.calls_filter')}</span>
    <select class="input" bind:value={filter}>
      {#each CALL_FILTERS as f (f)}<option value={f}>{t(`fritzbox.filter.${f}`)}</option>{/each}
    </select>
  </label>
  <label class="field">
    <span>{t('fritzbox.calls_count')}</span>
    <input class="input" type="number" min="5" max="50" bind:value={limit} required />
  </label>
  <label><input type="checkbox" bind:checked={showNumbers} /> {t('fritzbox.show_numbers')}</label>
  <button class="btn btn-primary" type="submit">{t('fritzbox.save')}</button>
</form>

<style>
  form {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
  }
</style>
