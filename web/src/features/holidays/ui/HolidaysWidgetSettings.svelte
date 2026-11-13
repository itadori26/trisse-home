<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseHolidaysConfig, STATES, stateName } from '../domain/holidays';

  /** Widget settings: state and number of entries. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseHolidaysConfig(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field"
    ><span>{t('holidays.state')}</span><select class="input" bind:value={form.state}>
      {#each STATES as s (s)}<option value={s}>{stateName(s)}</option>{/each}
    </select></label
  >
  <label class="field"
    ><span>{t('holidays.count', { n: form.count })}</span><input
      type="range"
      min="1"
      max="20"
      bind:value={form.count}
    /></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
