<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseWidgetConfig } from '../domain/media';

  /** Props of the media widget settings. */
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
  <label class="check"
    ><input type="checkbox" bind:checked={form.showSessions} /> {t('media.now_playing')}</label
  >
  <label class="check"
    ><input type="checkbox" bind:checked={form.showRecent} /> {t('media.recently_added')}</label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .check {
    display: block;
  }
</style>
