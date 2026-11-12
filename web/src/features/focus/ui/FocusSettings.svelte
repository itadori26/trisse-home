<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type FocusConfig, MAX_MINUTES, parseFocusConfig } from '../domain/focus';
  import type { FocusUseCases } from '../usecase/focusUseCases';

  /** Props of the focus timer settings (#58). */
  interface Props {
    config: Record<string, unknown>;
    focus: FocusUseCases;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, focus, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<FocusConfig>(parseFocusConfig(config));
  let hint = $state('');

  async function toggleNotify() {
    if (!form.notify) return;
    // The permission prompt needs this click; ask right away.
    const p = await focus.notifier.request();
    if (p !== 'granted') {
      form.notify = false;
      hint = t(p === 'unsupported' ? 'focus.notify_unsupported' : 'focus.notify_denied');
    } else hint = '';
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    await onsave({ ...config, ...parseFocusConfig({ ...form }) });
  }
</script>

<form onsubmit={submit}>
  <div class="row">
    <label class="field">
      <span>{t('focus.work_minutes')}</span>
      <input class="input" type="number" min="1" max={MAX_MINUTES} bind:value={form.work} />
    </label>
    <label class="field">
      <span>{t('focus.short_minutes')}</span>
      <input class="input" type="number" min="1" max={MAX_MINUTES} bind:value={form.short} />
    </label>
    <label class="field">
      <span>{t('focus.long_minutes')}</span>
      <input class="input" type="number" min="1" max={MAX_MINUTES} bind:value={form.long} />
    </label>
  </div>
  <label class="field">
    <span>{t('focus.long_every')}</span>
    <input class="input" type="number" min="1" max="12" bind:value={form.longEvery} />
  </label>
  <label
    ><input type="checkbox" bind:checked={form.notify} onchange={toggleNotify} /> {t('focus.notify')}</label
  >
  {#if hint}<p class="muted" role="status">{hint}</p>{/if}
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row .field {
    flex: 1 1 6rem;
  }
</style>
