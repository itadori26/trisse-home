<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseWidgetConfig } from '../domain/releases';

  /** Props of the new-releases widget settings. */
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
  <label class="field"
    ><span>{t('releases.weeks')}</span><input
      class="input"
      type="number"
      min="1"
      max="12"
      bind:value={form.weeks}
    /></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
