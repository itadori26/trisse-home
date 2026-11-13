<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { dnsProviders, parseWidgetConfig, providerName } from '../domain/adguard';

  /** Props of the DNS filter widget settings: which provider it shows (#83). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseWidgetConfig(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field">
    <span>{t('adguard.provider')}</span>
    <select class="input" bind:value={form.provider}>
      {#each dnsProviders as p (p)}<option value={p}>{providerName(p)}</option>{/each}
    </select>
  </label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
