<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseReadLaterConfig } from '../domain/readlater';

  /** Widget settings: how many articles are listed. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let count = $state(parseReadLaterConfig(config).count);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ count });
  }}
>
  <label class="field"
    ><span>{t('readlater.count', { n: count })}</span><input
      type="range"
      min="1"
      max="30"
      bind:value={count}
    /></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
