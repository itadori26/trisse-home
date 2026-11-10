<script lang="ts">
  import type { SearchUseCases } from '../usecase/searchUseCases';

  /** Props of the search bar widget. */
  interface Props {
    search: SearchUseCases;
    config: Record<string, unknown>;
  }
  let { search, config }: Props = $props();

  let query = $state('');
  const newTab = $derived(config['newTab'] === true);

  function submit(e: SubmitEvent) {
    e.preventDefault();
    const target = search.targetFor(query);
    if (!target) return;
    if (newTab) window.open(target, '_blank', 'noopener');
    else window.location.assign(target);
    query = '';
  }
</script>

<form class="search" role="search" onsubmit={submit}>
  <input
    class="input"
    type="search"
    name="q"
    bind:value={query}
    placeholder="Search — !bangs like !w work"
    aria-label="Search"
    autocomplete="off"
  />
</form>

<style>
  .search {
    height: 100%;
    display: flex;
    align-items: center;
  }
  .input {
    font-size: 1.15rem;
    padding: var(--space-3) var(--space-4);
    border-radius: 999px;
    background: var(--color-surface);
    box-shadow: var(--shadow);
  }
</style>
