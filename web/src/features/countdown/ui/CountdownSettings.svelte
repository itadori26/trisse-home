<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type CountdownConfig, MAX_GRACE_DAYS, parseCountdownConfig } from '../domain/countdown';

  /** Props of the countdown widget settings (#54). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<CountdownConfig>(parseCountdownConfig(config));

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    await onsave({ ...config, ...parseCountdownConfig({ ...form }) });
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>{t('countdown.highlight_days')}</span>
    <input class="input" type="number" min="0" max="365" bind:value={form.highlightDays} />
  </label>
  <label class="field">
    <span>{t('countdown.grace_days')}</span>
    <input class="input" type="number" min="0" max={MAX_GRACE_DAYS} bind:value={form.graceDays} />
  </label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
