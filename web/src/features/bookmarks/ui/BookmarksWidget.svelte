<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { type Bookmark, iconSrc, linkProblem, parseBookmarksConfig } from '../domain/bookmark';
  import type { BookmarkList, BookmarksUseCases } from '../usecase/bookmarksUseCases';
  import BookmarkForm from './BookmarkForm.svelte';

  /** Props of the bookmarks widget. */
  interface Props {
    bookmarks: BookmarksUseCases;
    config: Record<string, unknown>;
  }
  let { bookmarks, config }: Props = $props();

  const cfg = $derived(parseBookmarksConfig(config));
  let data = $state<BookmarkList | null>(null);
  let query = $state('');
  let error = $state('');
  let editing = $state<Bookmark | 'new' | null>(null);
  // "Broken links only" (#72); offered once a check found a broken link.
  let brokenOnly = $state(false);
  const anyBroken = $derived(data?.bookmarks.some((b) => b.link.state === 'broken') ?? false);

  async function load() {
    try {
      data = await bookmarks.list({ folderId: cfg.folderId, tag: cfg.tag, query, broken: brokenOnly });
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void cfg;
    void query;
    void brokenOnly;
    load();
  });

  async function remove(b: Bookmark) {
    if (!confirm(t('common.confirm_delete', { name: b.title }))) return;
    await bookmarks.delete(b.id).catch((err) => (error = errorMessage(err)));
    editing = null;
    load();
  }
</script>

<div class="toolbar">
  <input
    class="input"
    type="search"
    placeholder={t('bookmarks.filter')}
    bind:value={query}
    aria-label={t('bookmarks.filter_bookmarks')}
  />
  {#if anyBroken || brokenOnly}
    <button
      class="btn btn-icon"
      type="button"
      aria-pressed={brokenOnly}
      title={t('bookmarks.broken_only')}
      aria-label={t('bookmarks.broken_only')}
      onclick={() => (brokenOnly = !brokenOnly)}>⚠</button
    >
  {/if}
  <button
    class="btn btn-icon"
    type="button"
    aria-label={t('bookmarks.add_bookmark')}
    onclick={() => (editing = 'new')}>＋</button
  >
</div>
{#if error}<p class="error">{error}</p>{/if}
{#if data}
  <ul class="list">
    {#each data.bookmarks as b (b.id)}
      {@const icon = iconSrc(b)}
      {@const problem = linkProblem(b.link, intlLocale())}
      <li>
        <a href={b.url} target="_blank" rel="noopener noreferrer" title={b.description || b.url}>
          {#if icon}<img
              src={icon}
              alt=""
              width="16"
              height="16"
              loading="lazy"
              referrerpolicy="no-referrer"
              onerror={(e) => ((e.currentTarget as HTMLImageElement).style.visibility = 'hidden')}
            />{/if}
          <span class="title">{b.title}</span>
        </a>
        {#if problem}<span
            class="link-state"
            class:broken={b.link.state === 'broken'}
            title={problem}
            aria-label={problem}
            role="img">{b.link.state === 'broken' ? '⚠' : '↪'}</span
          >{/if}
        {#each b.tags as t (t)}<span class="tag">#{t}</span>{/each}
        <button
          class="btn btn-icon edit"
          type="button"
          aria-label={t('bookmarks.edit', { title: b.title })}
          onclick={() => (editing = b)}>✎</button
        >
      </li>
    {:else}
      <li class="muted">
        {brokenOnly ? t('bookmarks.no_broken') : query ? t('bookmarks.none_match') : t('bookmarks.none_yet')}
      </li>
    {/each}
  </ul>
{/if}

{#if editing && data}
  <Modal
    title={editing === 'new' ? t('bookmarks.add_bookmark') : t('bookmarks.edit_bookmark')}
    onclose={() => (editing = null)}
  >
    <BookmarkForm
      {bookmarks}
      folders={data.folders}
      bookmark={editing === 'new' ? undefined : editing}
      folderId={cfg.folderId === 'none' ? '' : cfg.folderId}
      onsaved={() => {
        editing = null;
        load();
      }}
    />
    {#if editing !== 'new'}
      {@const b = editing}
      <p>
        <button class="btn btn-danger" type="button" onclick={() => remove(b)}
          >{t('bookmarks.delete_bookmark')}</button
        >
      </p>
    {/if}
  </Modal>
{/if}

<style>
  .toolbar {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 3px 0;
  }
  a {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
    color: inherit;
    text-decoration: none;
  }
  a:hover .title {
    text-decoration: underline;
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .link-state {
    font-size: 0.8em;
    color: var(--color-text-muted);
  }
  .link-state.broken {
    color: var(--color-danger);
  }
  .tag {
    font-size: 0.75em;
    color: var(--color-text-muted);
  }
  .edit {
    margin-left: auto;
    opacity: 0;
  }
  li:hover .edit,
  .edit:focus {
    opacity: 1;
  }
</style>
