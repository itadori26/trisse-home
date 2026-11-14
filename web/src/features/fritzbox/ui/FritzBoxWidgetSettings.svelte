<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { internetOptions } from '../domain/fritzbox';

  /** Settings of the FRITZ!Box internet widget: which values it shows (#168). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  const initial = internetOptions(config);
  let showIp = $state(initial.showIp);
  let showRates = $state(initial.showRates);
  let showUptime = $state(initial.showUptime);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...config, showIp, showRates, showUptime });
  }}
>
  <label><input type="checkbox" bind:checked={showUptime} /> {t('fritzbox.show_uptime')}</label>
  <label><input type="checkbox" bind:checked={showIp} /> {t('fritzbox.show_ip')}</label>
  <label><input type="checkbox" bind:checked={showRates} /> {t('fritzbox.show_rates')}</label>
  <button class="btn btn-primary" type="submit">{t('fritzbox.save')}</button>
</form>

<style>
  form {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
  }
</style>
