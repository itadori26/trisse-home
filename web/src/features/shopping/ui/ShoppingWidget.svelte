<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { type ShoppingItem, withDone } from '../domain/shopping';
  import type { ShoppingUseCases } from '../usecase/shoppingUseCases';

  /** Props of the shopping list widget. */
  interface Props {
    shopping: ShoppingUseCases;
  }
  let { shopping }: Props = $props();

  const POLL_MS = 60_000;
  let items = $state<ShoppingItem[]>([]);
  let loaded = $state(false);
  let draft = $state('');
  let error = $state('');
  let unconfigured = $state(false);
  let busy = $state(false);
  const doneCount = $derived(items.filter((i) => i.done).length);

  const load = () =>
    shopping
      .items()
      .then((l) => ((items = l), (loaded = true), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
      });

  $effect(() => {
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  async function add(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    try {
      const it = await shopping.add(draft);
      if (it) items = [...items.filter((i) => !i.done), it, ...items.filter((i) => i.done)];
      draft = ''; // only cleared once the server confirmed the item
      error = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function toggle(it: ShoppingItem) {
    const before = items;
    items = withDone(items, it.id, !it.done);
    try {
      await shopping.setDone(it.id, !it.done);
      error = '';
    } catch (err) {
      items = before;
      error = errorMessage(err);
    }
  }

  async function clear() {
    try {
      await shopping.clearCompleted();
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if unconfigured}
  <p class="muted">{t('shopping.connect_a_caldav_task_list_under')}</p>
{:else}
  <form class="add" onsubmit={add}>
    <input
      class="input"
      bind:value={draft}
      placeholder={t('shopping.add_item')}
      aria-label={t('shopping.new_item')}
      maxlength="200"
      disabled={!loaded}
    />
    <button class="btn" type="submit" disabled={busy || !draft.trim()}>{t('shopping.add')}</button>
  </form>
  {#if error}<p class="error small" role="alert">{error}</p>{/if}
  <ul class="items">
    {#each items as it (it.id)}
      <li class:done={it.done}>
        <label
          ><input type="checkbox" checked={it.done} onchange={() => toggle(it)} />
          <span>{it.summary}</span></label
        >
      </li>
    {:else}
      {#if loaded}<li class="muted">{t('shopping.the_list_is_empty')}</li>{/if}
    {/each}
  </ul>
  {#if doneCount}
    <button class="btn link" type="button" onclick={clear}
      >{t('shopping.remove_done', { doneCount: doneCount })}</button
    >
  {/if}
{/if}

<style>
  .add {
    display: flex;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .add .input {
    flex: 1;
    min-width: 0;
  }
  .items {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li label {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    padding: 1px 0;
    cursor: pointer;
  }
  .done span {
    text-decoration: line-through;
    color: var(--color-text-muted);
  }
  .link {
    margin-top: var(--space-1);
    font-size: 0.8em;
  }
  .small {
    font-size: 0.8em;
  }
</style>
