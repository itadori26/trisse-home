<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type CalendarFile, emptyRule, type WasteRule } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /** Props of the waste calendar settings. */
  interface Props {
    waste: WasteUseCases;
  }
  let { waste }: Props = $props();

  /** The source shown in the form; switching it changes nothing until saved or uploaded (#153). */
  let mode = $state<'url' | 'file'>('url');
  let url = $state('');
  let file = $state<CalendarFile | null>(null);
  let rules = $state<WasteRule[]>([]);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);
  let picker = $state<HTMLInputElement>();

  $effect(() => {
    waste.settings().then((s) => {
      url = s.url;
      file = s.file;
      rules = s.rules;
      mode = s.source === 'file' ? 'file' : 'url';
    });
  });

  const day = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short', year: 'numeric' });
  // The same 30-day rule as the server: ask for next year's file in time.
  const endsSoon = $derived(file !== null && file.to.getTime() - Date.now() < 30 * 86_400_000);

  async function run(
    action: () => Promise<{ url: string; file: CalendarFile | null; rules: WasteRule[] }>,
    done: string,
  ) {
    error = message = '';
    busy = true;
    try {
      const s = await action();
      url = s.url;
      file = s.file;
      rules = s.rules;
      message = done;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function submit(e: SubmitEvent) {
    e.preventDefault();
    // In file mode only the rules are saved; the uploaded file stays.
    void run(() => waste.saveSettings(mode === 'url' ? url : '', rules), t('waste.loaded_saved'));
  }

  function upload() {
    const chosen = picker?.files?.[0];
    if (!chosen) return;
    void run(() => waste.uploadCalendar(chosen), t('waste.file_uploaded')).finally(() => {
      if (picker) picker.value = '';
    });
  }

  function remove() {
    void run(() => waste.removeCalendar(), t('waste.file_removed'));
  }
</script>

<p class="muted">
  {t('waste.many_municipalities_offer_the_collection')} <code>webcal://</code>{t(
    'waste.types_are_recognised_automatically_restm',
  )}
</p>
<div class="modes" role="radiogroup" aria-label={t('waste.source')}>
  <label class:active={mode === 'url'}
    ><input type="radio" bind:group={mode} value="url" />{t('waste.source_url')}</label
  >
  <label class:active={mode === 'file'}
    ><input type="radio" bind:group={mode} value="file" />{t('waste.source_file')}</label
  >
</div>
{#if mode === 'file'}
  <section class="file" aria-label={t('waste.source_file')}>
    <p class="muted small">{t('waste.file_hint')}</p>
    {#if file}
      <dl>
        <dt>{t('waste.file_name')}</dt>
        <dd>{file.name}</dd>
        <dt>{t('waste.file_uploaded_at')}</dt>
        <dd>{day(file.uploadedAt)}</dd>
        <dt>{t('waste.file_dates')}</dt>
        <dd>{t('waste.file_range', { n: file.events, from: day(file.from), to: day(file.to) })}</dd>
      </dl>
      {#if endsSoon}<p class="warning" role="status">
          {t('waste.file_ends_soon', { date: day(file.to) })}
        </p>{/if}
    {/if}
    <div class="row">
      <label class="btn"
        >{file ? t('waste.file_replace') : t('waste.file_choose')}<input
          bind:this={picker}
          class="file-input"
          type="file"
          accept=".ics,text/calendar"
          disabled={busy}
          onchange={upload}
        /></label
      >
      {#if file}<button class="btn" type="button" disabled={busy} onclick={remove}
          >{t('waste.file_remove')}</button
        >{/if}
    </div>
  </section>
{/if}
<form onsubmit={submit}>
  {#if mode === 'url'}
    <label class="field"
      ><span>{t('waste.ics_link')}</span><input
        class="input"
        bind:value={url}
        placeholder="https://…/abfuhrkalender.ics"
        required
      /></label
    >
  {/if}
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
  <button class="btn btn-primary" type="submit" disabled={busy || (mode === 'file' && !file)}
    >{busy ? t('common.loading') : mode === 'url' ? t('common.test_save') : t('waste.save_rules')}</button
  >
</form>

<style>
  .modes {
    display: inline-flex;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    overflow: hidden;
    margin-bottom: var(--space-2);
  }
  .modes label {
    padding: var(--space-1) var(--space-3);
    cursor: pointer;
  }
  .modes label.active {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .modes label:has(:focus-visible) {
    outline: 2px solid var(--color-accent);
    outline-offset: -2px;
  }
  .modes input,
  .file-input {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
  }
  .row label:has(:focus-visible) {
    outline: 2px solid var(--color-accent);
  }
  .file dl {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 2px var(--space-3);
    margin: var(--space-2) 0;
  }
  .file dt {
    color: var(--color-text-muted);
  }
  .file dd {
    margin: 0;
    overflow-wrap: anywhere;
  }
  .row {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .warning {
    color: var(--color-warning);
  }
  .small {
    font-size: 0.85em;
  }
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
