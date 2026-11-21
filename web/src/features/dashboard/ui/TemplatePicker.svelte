<script lang="ts">
  import { t, type MessageKey } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { browserLocation } from '../../../shared/places/places';
  import { type BuiltinTemplate, previewRows, type Tab, type TemplateLocation } from '../domain/dashboard';
  import type { DashboardUseCases } from '../usecase/dashboardUseCases';

  /** A place found by the search, for the location step. */
  interface FoundPlace {
    label: string;
    lat: number;
    lon: number;
  }

  /**
   * Props of the template choice on first start (#589): a card per template
   * with a small preview, then the location when the template needs one.
   */
  interface Props {
    uc: DashboardUseCases;
    widgetName: (kind: string) => string;
    /** Place search (the weather geocoder); omitted when it is off. */
    searchPlaces?: (query: string) => Promise<FoundPlace[]>;
    /**
     * The dashboard holds only the example tabs of the setup (#670): a
     * template or "start empty" replaces them, closing the dialog keeps them.
     */
    sample?: boolean;
    /** A place chosen just before (#671); the location step is not asked then. */
    location?: TemplateLocation;
    oncreated: (tab: Tab) => void;
    /** The example tabs were removed for an empty start (#670). */
    onemptied?: () => void;
    onclose: () => void;
  }
  let {
    uc,
    widgetName,
    searchPlaces,
    sample = false,
    location,
    oncreated,
    onemptied,
    onclose,
  }: Props = $props();

  let list = $state<BuiltinTemplate[]>([]);
  let chosen = $state('');
  let step = $state<'choose' | 'location'>('choose');
  let query = $state('');
  let results = $state<FoundPlace[]>([]);
  let busy = $state(false);
  let error = $state('');

  $effect(() => {
    uc.listBuiltinTemplates()
      .then((l) => {
        list = l;
        chosen = l[0]?.id ?? '';
      })
      .catch((err) => (error = errorMessage(err)));
  });

  const titleOf = (tpl: BuiltinTemplate) => {
    const key = `templates.${tpl.id}` as MessageKey;
    const text = t(key);
    return text === key ? tpl.title : text;
  };
  const current = $derived(list.find((x) => x.id === chosen));

  async function create(location?: TemplateLocation) {
    if (!current) return;
    busy = true;
    error = '';
    try {
      const tab = await uc.createTabFromBuiltin(current.id, titleOf(current), location);
      if (sample) await uc.discardSample();
      oncreated(tab);
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  /** Settles the example tabs, then closes: removed for an empty start, else kept. */
  async function settle(discard: boolean) {
    busy = true;
    error = '';
    try {
      if (discard) {
        await uc.discardSample();
        onemptied?.();
      } else {
        await uc.keepSample();
      }
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function startEmpty() {
    if (sample) void settle(true);
    else onclose();
  }

  // Closing the dialog (Escape, close button) never removes anything.
  function dismiss() {
    if (sample) void settle(false);
    else onclose();
  }

  function next() {
    if (current?.needs.includes('location') && !location) step = 'location';
    else void create(current?.needs.includes('location') ? location : undefined);
  }

  async function find(e: SubmitEvent) {
    e.preventDefault();
    if (!searchPlaces) return;
    error = '';
    try {
      results = await searchPlaces(query);
      if (results.length === 0) error = t('templates.no_results');
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function locate() {
    error = '';
    try {
      const at = await browserLocation();
      await create({ ...at, label: '' });
    } catch {
      error = t('templates.location_denied');
    }
  }
</script>

<Modal
  title={step === 'choose' ? t('templates.title') : t('templates.location_title')}
  size="wide"
  onclose={dismiss}
>
  {#if step === 'choose'}
    <p class="muted">{t(sample ? 'templates.intro_sample' : 'templates.intro')}</p>
    <div class="cards" role="radiogroup" aria-label={t('templates.title')}>
      {#each list as tpl (tpl.id)}
        <label class="card" class:chosen={chosen === tpl.id}>
          <input type="radio" name="template" value={tpl.id} bind:group={chosen} />
          <strong>{titleOf(tpl)}</strong>
          <!-- A miniature of the 12-column grid; decorative, the list below names the widgets. -->
          <div
            class="preview"
            style="grid-template-rows: repeat({previewRows(tpl)}, 6px)"
            role="img"
            aria-label={t('templates.preview', { name: titleOf(tpl) })}
          >
            {#each tpl.widgets as w, i (i)}
              <span style="grid-column: {w.x + 1} / span {w.w}; grid-row: {w.y + 1} / span {w.h}"></span>
            {/each}
          </div>
          <span class="muted small">{tpl.widgets.map((w) => widgetName(w.kind)).join(', ')}</span>
        </label>
      {/each}
    </div>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="actions">
      {#if sample}
        <button class="btn" type="button" disabled={busy} onclick={() => settle(false)}
          >{t('templates.keep_sample')}</button
        >
      {/if}
      <button class="btn" type="button" disabled={busy} onclick={startEmpty}>{t('templates.empty')}</button>
      <button class="btn btn-primary" type="button" disabled={!current || busy} onclick={next}
        >{t('templates.use')}</button
      >
    </div>
  {:else}
    <p class="muted">{t('templates.location_intro')}</p>
    {#if searchPlaces}
      <form class="row" onsubmit={find}>
        <input class="input grow" bind:value={query} aria-label={t('templates.location_search')} />
        <button class="btn" type="submit">{t('templates.search')}</button>
      </form>
      {#if results.length}
        <ul class="results">
          {#each results as r (`${r.lat},${r.lon}`)}
            <li>
              <button class="btn" type="button" disabled={busy} onclick={() => create({ ...r })}
                >{r.label}</button
              >
            </li>
          {/each}
        </ul>
      {/if}
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="actions">
      <button class="btn" type="button" disabled={busy} onclick={locate}>{t('templates.my_location')}</button>
      <button class="btn" type="button" disabled={busy} onclick={() => create()}
        >{t('templates.skip_location')}</button
      >
    </div>
  {/if}
</Modal>

<style>
  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: var(--space-2);
    margin: var(--space-3) 0;
  }
  .card {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    cursor: pointer;
  }
  .card.chosen {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 1px var(--color-accent);
  }
  .card input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }
  .card:focus-within {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .preview {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 2px;
  }
  .preview span {
    background: color-mix(in srgb, var(--color-accent) 35%, transparent);
    border-radius: 2px;
  }
  .small {
    font-size: 0.85rem;
  }
  .actions,
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    justify-content: flex-end;
  }
  .row {
    justify-content: flex-start;
    margin-bottom: var(--space-2);
  }
  .grow {
    flex: 1;
    min-width: 12rem;
  }
  .results {
    list-style: none;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
</style>
