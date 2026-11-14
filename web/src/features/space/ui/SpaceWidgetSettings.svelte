<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseSpaceConfig } from '../domain/space';

  /** Widget settings: what to show and how many launches. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseSpaceConfig(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field"
    ><span>{t('space.show')}</span><select class="input" bind:value={form.show}>
      <option value="both">{t('space.show_both')}</option>
      <option value="launches">{t('space.show_launches')}</option>
      <option value="picture">{t('space.show_picture')}</option>
    </select></label
  >
  {#if form.show !== 'picture'}
    <label class="field"
      ><span>{t('space.count', { n: form.count })}</span><input
        type="range"
        min="1"
        max="10"
        bind:value={form.count}
      /></label
    >
  {/if}
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
