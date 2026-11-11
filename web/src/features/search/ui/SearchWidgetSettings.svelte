<script lang="ts">
  import { t } from '../../../shared/i18n';
  /** Props of the per-widget search settings. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let newTab = $state(config['newTab'] === true);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ newTab });
  }}
>
  <label><input type="checkbox" bind:checked={newTab} /> {t('search.open_results_in_a_new_tab')}</label>
  <p class="muted">{t('search.the_search_engine_and_custom_shortcuts')}</p>
  <button class="btn btn-primary" type="submit">{t('search.save')}</button>
</form>
