<script lang="ts">
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

  run(() => search.settings());
</script>

{#if settings}
  <label class="field">
    <span>Search engine</span>
    <select
      class="input"
      value={settings.defaultProvider}
      onchange={(e) => run(() => search.setDefaultProvider(e.currentTarget.value))}
    >
      {#each settings.providers as p (p.id)}<option value={p.id}>{p.name}</option>{/each}
    </select>
  </label>
  <p class="muted">
    DuckDuckGo understands thousands of !bangs (e.g. <code>!w</code>, <code>!g</code>, <code>!osm</code>).
    Custom shortcuts below take precedence and work with every engine.
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
            aria-label="Delete shortcut !{s.keyword}"
            onclick={() => run(() => search.deleteShortcut(s.keyword))}>✕</button
          >
        </li>
      {/each}
    </ul>
  {/if}
  <form class="add" onsubmit={add}>
    <input class="input" bind:value={keyword} placeholder="!gh" aria-label="Keyword" required />
    <input class="input" bind:value={name} placeholder="GitHub" aria-label="Name" required />
    <input
      class="input"
      bind:value={urlTemplate}
      placeholder={'https://github.com/search?q={query}'}
      aria-label="URL with {'{query}'}"
      required
    />
    <button class="btn" type="submit">Add</button>
  </form>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .shortcuts {
    list-style: none;
    padding: 0;
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
