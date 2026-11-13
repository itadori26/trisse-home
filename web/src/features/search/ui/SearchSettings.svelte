<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { SearchSettings } from '../usecase/ports';
  import type { SearchUseCases } from '../usecase/searchUseCases';

  /** Props of the global search settings section. */
  interface Props {
    search: SearchUseCases;
  }
  let { search }: Props = $props();

  let settings = $state<SearchSettings | null>(null);
  let error = $state('');
  let keyword = $state('');
  let name = $state('');
  let urlTemplate = $state('');

  async function run(action: () => Promise<SearchSettings>) {
    error = '';
    try {
      settings = await action();
      return true;
    } catch (err) {
      error = errorMessage(err);
      return false;
    }
  }

  async function add(e: SubmitEvent) {
    e.preventDefault();
    if (await run(() => search.addShortcut({ keyword, name, urlTemplate })))
      keyword = name = urlTemplate = '';
  }

  // Options of the search field (#221), formerly options of the search widget.
  let newTab = $state(false);
  let todo = $state('t');
  let note = $state('n');
  let bookmark = $state('b');
  let barSaved = $state(false);
  function fill(s: SearchSettings) {
    newTab = s.bar.openInNewTab;
    ({ todo, note, bookmark } = s.bar.capturePrefixes);
  }
  async function saveBar(e: SubmitEvent) {
    e.preventDefault();
    barSaved = false;
    const options = { openInNewTab: newTab, capturePrefixes: { todo, note, bookmark } };
    if (await run(() => search.setBar(options))) {
      if (settings) fill(settings);
      barSaved = true;
    }
  }

  void run(() => search.settings()).then((ok) => ok && settings && fill(settings));
</script>

{#if settings}
  <label class="field">
    <span>{t('search.search_engine')}</span>
    <select
      class="input"
      value={settings.defaultProvider}
      onchange={(e) => run(() => search.setDefaultProvider(e.currentTarget.value))}
    >
      {#each settings.providers as p (p.id)}<option value={p.id}>{p.name}</option>{/each}
    </select>
  </label>
  <p class="muted">
    {t('search.duckduckgo_understands_thousands_of_bang')} <code>!w</code>, <code>!g</code>,
    <code>!osm</code>{t('search.custom_shortcuts_below_take_precedence_a')}
  </p>
  {#if settings.shortcuts.length}
    <ul class="shortcuts">
      {#each settings.shortcuts as s (s.keyword)}
        <li>
          <code>!{s.keyword}</code>
          {s.name}
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('search.delete_shortcut', { keyword: s.keyword })}
            onclick={() => run(() => search.deleteShortcut(s.keyword))}>✕</button
          >
        </li>
      {/each}
    </ul>
  {/if}
  <form class="add" onsubmit={add}>
    <input class="input" bind:value={keyword} placeholder="!gh" aria-label={t('search.keyword')} required />
    <input
      class="input"
      bind:value={name}
      placeholder={t('search.github')}
      aria-label={t('search.name')}
      required
    />
    <input
      class="input"
      bind:value={urlTemplate}
      placeholder={'https://github.com/search?q={query}'}
      aria-label={t('search.url_with', { value: '{query}' })}
      required
    />
    <button class="btn" type="submit">{t('search.add')}</button>
  </form>
  <form class="bar" onsubmit={saveBar}>
    <h4>{t('search.search_field')}</h4>
    <label><input type="checkbox" bind:checked={newTab} /> {t('search.open_results_in_a_new_tab')}</label>
    <fieldset>
      <legend>{t('quick.prefixes')}</legend>
      <p class="muted">{t('quick.prefixes_hint')}</p>
      <label class="field"
        ><span>{t('quick.prefix_todo')}</span><input
          class="input"
          bind:value={todo}
          maxlength="4"
          required
        /></label
      >
      <label class="field"
        ><span>{t('quick.prefix_note')}</span><input
          class="input"
          bind:value={note}
          maxlength="4"
          required
        /></label
      >
      <label class="field"
        ><span>{t('quick.prefix_bookmark')}</span><input
          class="input"
          bind:value={bookmark}
          maxlength="4"
          required
        /></label
      >
    </fieldset>
    <button class="btn btn-primary" type="submit">{t('search.save')}</button>
    {#if barSaved}<span class="muted" role="status">{t('search.saved')}</span>{/if}
  </form>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .shortcuts {
    list-style: none;
    padding: 0;
  }
  .bar {
    display: grid;
    gap: var(--space-2);
    margin-top: var(--space-4);
  }
  .bar h4 {
    margin: 0;
  }
  .add {
    display: grid;
    grid-template-columns: 5rem 1fr 2fr auto;
    gap: var(--space-2);
  }
  @media (max-width: 600px) {
    .add {
      grid-template-columns: 1fr;
    }
  }
</style>
