<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type HabitsConfig, parseHabitsConfig } from '../domain/habits';

  /** Props of the habit tracker widget settings (#106). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<HabitsConfig>(parseHabitsConfig(config));

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    await onsave({ ...config, ...form });
  }
</script>

<form onsubmit={submit}>
  <label><input type="checkbox" bind:checked={form.showHeatmap} /> {t('habits.show_heatmap')}</label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
