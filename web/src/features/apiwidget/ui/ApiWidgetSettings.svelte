<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Endpoint, intervalLabel } from '../domain/apiwidget';
  import type { ApiWidgetUseCases } from '../usecase/apiWidgetUseCases';
  import EndpointEditor from './EndpointEditor.svelte';

  /** Settings → JSON APIs: the configured endpoints (#48). */
  interface Props {
    apiwidget: ApiWidgetUseCases;
  }
  let { apiwidget }: Props = $props();

  let endpoints = $state<Endpoint[]>([]);
  /** The endpoint being edited: an id, 'new', or null. */
  let editing = $state<string | null>(null);
  let error = $state('');
  let message = $state('');

  const load = () =>
    apiwidget
      .endpoints()
      .then((list) => (endpoints = list))
      .catch((err) => (error = errorMessage(err)));
  $effect(() => {
    void load();
  });

  function saved() {
    editing = null;
    message = t('apiwidget.saved');
    void load();
  }

  async function remove(e: Endpoint) {
    if (!confirm(t('apiwidget.delete_confirm', { name: e.name }))) return;
    error = message = '';
    try {
      await apiwidget.remove(e.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<p class="muted">{t('apiwidget.intro')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}
<ul class="list">
  {#each endpoints as e (e.id)}
    <li>
      {#if editing === e.id}
        <EndpointEditor {apiwidget} stored={e} onsaved={saved} oncancel={() => (editing = null)} />
      {:else}
        <div class="item">
          <span class="what">
            <strong>{e.name}</strong>
            <span class="muted small"
              >{e.url} · {intervalLabel(e.intervalSeconds)} · {e.extractions.length} × JSONPath</span
            >
          </span>
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('apiwidget.edit', { name: e.name })}
            onclick={() => ((editing = e.id), (message = ''))}>✎</button
          >
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('apiwidget.delete', { name: e.name })}
            onclick={() => remove(e)}>✕</button
          >
        </div>
      {/if}
    </li>
  {:else}
    {#if editing !== 'new'}<li class="muted">{t('apiwidget.no_endpoints')}</li>{/if}
  {/each}
</ul>
{#if editing === 'new'}
  <EndpointEditor {apiwidget} onsaved={saved} oncancel={() => (editing = null)} />
{:else}
  <button class="btn btn-primary" type="button" onclick={() => ((editing = 'new'), (message = ''))}
    >{t('apiwidget.add')}</button
  >
{/if}

<style>
  .list {
    list-style: none;
    margin: 0 0 var(--space-3);
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .what {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .small {
    font-size: 0.8em;
    overflow-wrap: anywhere;
  }
</style>
