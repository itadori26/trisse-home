<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Article, parseReadLaterConfig } from '../domain/readlater';
  import type { ReadLaterUseCases } from '../usecase/readLaterUseCases';
  import ArticleReader from './ArticleReader.svelte';

  /** Unread articles with check-off and a field to save a link (#99). */
  interface Props {
    readlater: ReadLaterUseCases;
    config: Record<string, unknown>;
  }
  let { readlater, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseReadLaterConfig(config));

  let articles = $state<Article[]>([]);
  let url = $state('');
  let error = $state('');
  let busy = $state(false);
  let open = $state<Article | null>(null);

  const load = () =>
    readlater
      .list('unread')
      .then((a) => ((articles = a), (error = '')))
      .catch((err) => {
        error = errorMessage(err);
        throw err;
      });
  $effect(() => poll(load, 5 * 60_000));

  async function run(action: () => Promise<unknown>) {
    error = '';
    busy = true;
    try {
      await action();
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    run(() => readlater.save(url)).then(() => !error && (url = ''));
  }}
>
  <input
    class="input"
    type="url"
    bind:value={url}
    placeholder={t('readlater.save_placeholder')}
    aria-label={t('readlater.save_label')}
  />
</form>
{#if busy}<p class="muted small" role="status">{t('readlater.saving')}</p>{/if}
{#if error}<p class="error small" role="alert">{error}</p>{/if}
<ul class="list">
  {#each articles.slice(0, cfg.count) as a (a.id)}
    <li>
      <input
        type="checkbox"
        aria-label={t('readlater.mark_read', { title: a.title })}
        onchange={() => run(() => readlater.setStatus(a.id, 'read'))}
      />
      <button class="title" type="button" onclick={async () => (open = await readlater.open(a.id))}>
        <span class="name">{a.title}</span>
        <span class="muted small">{a.site} · {t('readlater.minutes', { n: a.readingMinutes })}</span>
      </button>
    </li>
  {:else}
    <li class="muted">{t('readlater.empty')}</li>
  {/each}
</ul>
{#if open}
  <ArticleReader
    article={open}
    onclose={() => ((open = null), void load())}
    onarchive={() => {
      const id = open?.id ?? '';
      open = null;
      void run(() => readlater.setStatus(id, 'archived'));
    }}
  />
{/if}

<style>
  .list {
    list-style: none;
    margin: var(--space-2) 0 0;
    padding: 0;
    display: grid;
    gap: 4px;
  }
  li {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
  }
  .title {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    text-align: left;
    background: none;
    border: 0;
    padding: 0;
    color: inherit;
    font: inherit;
    cursor: pointer;
  }
  .name,
  .title .small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
  }
</style>
