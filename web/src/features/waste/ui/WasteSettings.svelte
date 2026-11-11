<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { emptyRule, type WasteRule } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /** Props of the waste calendar settings. */
  interface Props {
    waste: WasteUseCases;
  }
  let { waste }: Props = $props();

  let url = $state('');
  let rules = $state<WasteRule[]>([]);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    waste.settings().then((s) => ((url = s.url), (rules = s.rules)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await waste.saveSettings(url, rules);
      rules = s.rules;
      message = t('waste.loaded_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  {t('waste.many_municipalities_offer_the_collection')} <code>webcal://</code>{t(
    'waste.types_are_recognised_automatically_restm',
  )}
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>{t('waste.ics_link')}</span><input
      class="input"
      bind:value={url}
      placeholder="https://…/abfuhrkalender.ics"
      required
    /></label
  >
  <fieldset>
    <legend>{t('waste.rules_optional')}</legend>
    {#each rules as r, i (i)}
      <div class="rule">
        <input
          class="input"
          bind:value={r.match}
          placeholder={t('waste.contains')}
          aria-label={t('waste.match_text')}
          maxlength="40"
        />
        <input
          class="input"
          bind:value={r.label}
          placeholder={t('waste.label')}
          aria-label={t('waste.label')}
          maxlength="40"
        />
        <input
          class="input icon"
          bind:value={r.icon}
          placeholder="🗑️"
          aria-label={t('waste.icon')}
          maxlength="8"
        />
        <input
          type="color"
          value={r.color || '#64748b'}
          oninput={(e) => (r.color = e.currentTarget.value)}
          aria-label={t('waste.color')}
        />
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('waste.remove_rule')}
          onclick={() => rules.splice(i, 1)}>✕</button
        >
      </div>
    {/each}
    <button class="btn" type="button" onclick={() => rules.push(emptyRule())} disabled={rules.length >= 20}
      >{t('waste.add_rule')}</button
    >
  </fieldset>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.loading') : t('common.test_save')}</button
  >
</form>

<style>
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    margin: var(--space-3) 0;
    display: grid;
    gap: var(--space-1);
  }
  .rule {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  .rule .input {
    flex: 1;
    min-width: 0;
  }
  .rule .icon {
    flex: 0 0 3.5rem;
  }
</style>
