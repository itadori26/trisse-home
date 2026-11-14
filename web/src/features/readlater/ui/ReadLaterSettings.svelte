<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Article, type ArticleStatus, formatBytes, STATUSES } from '../domain/readlater';
  import type { ReadLaterUseCases } from '../usecase/readLaterUseCases';
  import ArticleReader from './ArticleReader.svelte';

  /** Settings → Read later: all articles by status, archive and delete (#99). */
  interface Props {
    readlater: ReadLaterUseCases;
  }
  let { readlater }: Props = $props();

  let status = $state<ArticleStatus>('unread');
  let articles = $state<Article[]>([]);
  let url = $state('');
  let error = $state('');
  let open = $state<Article | null>(null);

  const load = () =>
    readlater
      .list(status)
      .then((a) => (articles = a))
      .catch((err) => (error = errorMessage(err)));
  $effect(() => {
    void status;
    void load();
  });

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
  const total = $derived(articles.reduce((s, a) => s + a.bytes, 0));
</script>

<p class="muted">{t('readlater.intro')}</p>
<form
  class="save"
  onsubmit={(e) => {
    e.preventDefault();
    run(() => readlater.save(url)).then(() => !error && (url = ''));
  }}
>
  <input
    class="input"
    type="url"
    bind:value={url}
    placeholder="https://…"
    aria-label={t('readlater.save_label')}
    required
  />
  <button class="btn btn-primary" type="submit">{t('readlater.save')}</button>
</form>
<div class="tabs" role="radiogroup" aria-label={t('readlater.status')}>
  {#each STATUSES as s (s)}
    <label class:active={status === s}
      ><input type="radio" bind:group={status} value={s} />{t(`readlater.status_${s}`)}</label
    >
  {/each}
</div>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<ul class="list">
  {#each articles as a (a.id)}
    <li>
      <button class="title" type="button" onclick={async () => (open = await readlater.open(a.id))}>
        <strong>{a.title}</strong>
        <span class="muted small"
          >{a.site} · {t('readlater.minutes', { n: a.readingMinutes })} · {formatBytes(
            a.bytes,
            intlLocale(),
          )}</span
        >
        {#if a.error}<span class="error small">{a.error}</span>{/if}
      </button>
      {#if a.status !== 'archived'}
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('readlater.archive_item', { title: a.title })}
          onclick={() => run(() => readlater.setStatus(a.id, 'archived'))}>🗄</button
        >
      {:else}
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('readlater.restore_item', { title: a.title })}
          onclick={() => run(() => readlater.setStatus(a.id, 'unread'))}>↩</button
        >
      {/if}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('readlater.delete_item', { title: a.title })}
        onclick={() =>
          confirm(t('readlater.delete_confirm', { title: a.title })) && run(() => readlater.remove(a.id))}
        >✕</button
      >
    </li>
  {:else}
    <li class="muted">{t('readlater.empty')}</li>
  {/each}
</ul>
{#if articles.length}<p class="muted small">
    {t('readlater.storage', { size: formatBytes(total, intlLocale()) })}
  </p>{/if}
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
  .save {
    display: flex;
    gap: var(--space-2);
  }
  .save .input {
    flex: 1;
    min-width: 0;
  }
  .tabs {
    display: inline-flex;
    margin: var(--space-3) 0 var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    overflow: hidden;
  }
  .tabs label {
    padding: var(--space-1) var(--space-3);
    cursor: pointer;
  }
  .tabs label.active {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .tabs label:has(:focus-visible) {
    outline: 2px solid var(--color-accent);
    outline-offset: -2px;
  }
  .tabs input {
    position: absolute;
    opacity: 0;
    width: 1px;
    height: 1px;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  li {
    display: flex;
    align-items: flex-start;
    gap: var(--space-1);
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
    overflow-wrap: anywhere;
  }
  .small {
    font-size: 0.8em;
  }
</style>
