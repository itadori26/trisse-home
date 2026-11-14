<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    emptyExtraction,
    emptyInput,
    type Endpoint,
    type EndpointInput,
    formatValue,
    INTERVALS,
    intervalLabel,
    MAX_EXTRACTIONS,
    type Preview,
    toInput,
  } from '../domain/apiwidget';
  import type { ApiWidgetUseCases } from '../usecase/apiWidgetUseCases';

  /** Form of one endpoint with a test request (#48). */
  interface Props {
    apiwidget: ApiWidgetUseCases;
    /** The endpoint being edited, or undefined for a new one. */
    stored?: Endpoint;
    onsaved: (e: Endpoint) => void;
    oncancel: () => void;
  }
  let { apiwidget, stored, onsaved, oncancel }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<EndpointInput>(stored ? toInput(stored) : emptyInput());
  let preview = $state<Preview | null>(null);
  let error = $state('');
  let busy = $state<'' | 'test' | 'save'>('');

  // A number input bound to number | null: empty means no threshold.
  const num = (v: string) => (v.trim() === '' ? null : Number(v));

  async function test() {
    busy = 'test';
    error = '';
    try {
      preview = await apiwidget.preview(form, stored);
    } catch (err) {
      preview = null;
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    busy = 'save';
    error = '';
    try {
      onsaved(await apiwidget.save(form, stored));
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }
</script>

<form class="editor" onsubmit={save}>
  <div class="grid2">
    <label class="field"
      ><span>{t('apiwidget.name')}</span><input
        class="input"
        bind:value={form.name}
        maxlength="60"
        required
      /></label
    >
    <label class="field"
      ><span>{t('apiwidget.interval')}</span><select class="input" bind:value={form.intervalSeconds}>
        {#each INTERVALS as s (s)}<option value={s}>{intervalLabel(s)}</option>{/each}
      </select></label
    >
  </div>
  <label class="field"
    ><span>{t('apiwidget.url')}</span><input
      class="input"
      type="url"
      bind:value={form.url}
      placeholder="http://192.168.1.10:8080/api/status"
      required
    /></label
  >
  <div class="grid2">
    <label class="field"
      ><span>{t('apiwidget.header_name')}</span><input
        class="input"
        bind:value={form.headerName}
        placeholder="X-Api-Key"
        autocomplete="off"
      /></label
    >
    {#if form.headerName.trim()}
      <label class="field"
        ><span>{t('apiwidget.header_value')}</span><input
          class="input"
          type="password"
          bind:value={form.headerValue}
          autocomplete="new-password"
          placeholder={stored?.hasHeader ? t('apiwidget.header_stored') : ''}
        /></label
      >
    {/if}
  </div>
  <p class="muted small">{t('apiwidget.header_hint')}</p>
  <label class="check"
    ><input type="checkbox" bind:checked={form.allowPublic} /> {t('apiwidget.allow_public')}</label
  >
  <p class="muted small">{t('apiwidget.allow_public_hint')}</p>

  <fieldset>
    <legend>{t('apiwidget.values')}</legend>
    <p class="muted small">{t('apiwidget.path_hint')}</p>
    {#each form.extractions as x, i (i)}
      <div class="value">
        <input
          class="input"
          bind:value={x.label}
          placeholder={t('apiwidget.label')}
          aria-label={t('apiwidget.label')}
          maxlength="40"
        />
        <input
          class="input path"
          bind:value={x.path}
          placeholder="$.data.value"
          aria-label={t('apiwidget.path')}
          maxlength="200"
        />
        <input
          class="input unit"
          bind:value={x.unit}
          placeholder={t('apiwidget.unit')}
          aria-label={t('apiwidget.unit')}
          maxlength="12"
        />
        <input
          class="input num"
          type="number"
          step="any"
          value={x.warn ?? ''}
          oninput={(e) => (x.warn = num(e.currentTarget.value))}
          placeholder={t('apiwidget.warn')}
          aria-label={t('apiwidget.warn')}
        />
        <input
          class="input num"
          type="number"
          step="any"
          value={x.crit ?? ''}
          oninput={(e) => (x.crit = num(e.currentTarget.value))}
          placeholder={t('apiwidget.crit')}
          aria-label={t('apiwidget.crit')}
        />
        <label class="check small"
          ><input type="checkbox" bind:checked={x.below} /> {t('apiwidget.below')}</label
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('apiwidget.remove_value')}
          disabled={form.extractions.length === 1}
          onclick={() => form.extractions.splice(i, 1)}>✕</button
        >
      </div>
    {/each}
    <button
      class="btn"
      type="button"
      disabled={form.extractions.length >= MAX_EXTRACTIONS}
      onclick={() => form.extractions.push(emptyExtraction())}>{t('apiwidget.add_value')}</button
    >
  </fieldset>

  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if preview}
    <section class="preview" aria-label={t('apiwidget.test')}>
      {#if preview.error}<p class="error" role="alert">{preview.error}</p>{/if}
      <dl>
        {#each preview.values as v, i (i)}
          <div class="row {v.status}">
            <dt>{v.label}</dt>
            <dd>{v.error ? v.error : formatValue(v, intlLocale())}</dd>
          </div>
        {/each}
      </dl>
      {#if preview.body}
        <details>
          <summary>{t('apiwidget.response')}</summary>
          <pre>{preview.body}</pre>
        </details>
      {/if}
    </section>
  {/if}
  <div class="actions">
    <button class="btn" type="button" disabled={busy !== ''} onclick={test}
      >{busy === 'test' ? t('apiwidget.testing') : t('apiwidget.test')}</button
    >
    <button class="btn btn-primary" type="submit" disabled={busy !== ''}>{t('apiwidget.save')}</button>
    <button class="btn" type="button" onclick={oncancel}>{t('apiwidget.cancel')}</button>
  </div>
</form>

<style>
  .editor {
    display: grid;
    gap: var(--space-2);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius);
  }
  .grid2 {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
    gap: var(--space-2);
  }
  .check {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    display: grid;
    gap: var(--space-1);
    min-width: 0;
  }
  .value {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    align-items: center;
    padding-bottom: var(--space-1);
    border-bottom: 1px dashed var(--color-border);
  }
  .value .input {
    flex: 1 1 8rem;
    min-width: 0;
  }
  .value .path {
    flex: 2 1 12rem;
    font-family: var(--font-mono, monospace);
  }
  .value .unit,
  .value .num {
    flex: 0 1 6rem;
  }
  .preview dl {
    margin: 0;
    display: grid;
    gap: 2px;
  }
  .row {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
    border-left: 3px solid transparent;
    padding-left: var(--space-1);
  }
  .row.warning {
    border-left-color: var(--color-warning);
  }
  .row.critical {
    border-left-color: var(--color-danger);
  }
  dd {
    margin: 0;
    text-align: right;
    overflow-wrap: anywhere;
  }
  pre {
    max-height: 16rem;
    overflow: auto;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    font-size: 0.8em;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .small {
    font-size: 0.85em;
    margin: 0;
  }
</style>
