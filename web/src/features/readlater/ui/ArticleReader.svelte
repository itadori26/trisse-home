<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { Article } from '../domain/readlater';

  /**
   * Reader view of a saved article (#99). The copy was sanitised on the
   * server (allowlist rebuild); the app's CSP blocks inline scripts and
   * foreign images on top of that.
   */
  interface Props {
    article: Article;
    onclose: () => void;
    onarchive: () => void;
  }
  let { article, onclose, onarchive }: Props = $props();
  const saved = $derived(
    article.savedAt.toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short', year: 'numeric' }),
  );
</script>

<Modal title={article.title} {onclose}>
  <p class="meta muted">
    {article.site} · {t('readlater.minutes', { n: article.readingMinutes })} · {t('readlater.saved_on', {
      date: saved,
    })}
    · <a href={article.url} target="_blank" rel="noopener noreferrer">{t('readlater.original')}</a>
  </p>
  {#if article.content}
    <!-- Sanitised on the server: only allowlisted elements, http(s) links, stored images. -->
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    <article class="reader">{@html article.content}</article>
  {:else}
    <p class="muted">{article.error || t('readlater.no_copy')}</p>
  {/if}
  <p class="actions">
    {#if article.status !== 'archived'}<button class="btn" type="button" onclick={onarchive}
        >{t('readlater.archive')}</button
      >{/if}
  </p>
</Modal>

<style>
  .meta {
    margin-top: 0;
    font-size: 0.85em;
  }
  .reader {
    max-width: 42rem;
    line-height: 1.6;
    font-size: 1.05em;
    overflow-wrap: anywhere;
  }
  .reader :global(img) {
    max-width: 100%;
    height: auto;
    border-radius: var(--radius-small);
  }
  .reader :global(pre) {
    overflow-x: auto;
    padding: var(--space-2);
    background: var(--color-surface-2);
    border-radius: var(--radius-small);
  }
  .reader :global(blockquote) {
    margin: var(--space-2) 0;
    padding-left: var(--space-3);
    border-left: 3px solid var(--color-border);
    color: var(--color-text-muted);
  }
  .reader :global(table) {
    border-collapse: collapse;
    display: block;
    overflow-x: auto;
  }
  .reader :global(td),
  .reader :global(th) {
    border: 1px solid var(--color-border);
    padding: 2px var(--space-1);
  }
  .reader :global(figcaption) {
    font-size: 0.85em;
    color: var(--color-text-muted);
  }
</style>
