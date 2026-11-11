<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { type Bookmark, faviconUrl, parseBookmarksConfig } from '../domain/bookmark';
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

  async function load() {
    try {
      data = await bookmarks.list({ folderId: cfg.folderId, tag: cfg.tag, query });
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void cfg;
    void query;
    load();
  });

  async function remove(b: Bookmark) {
    if (!confirm(`Delete "${b.title}"?`)) return;
    await bookmarks.delete(b.id).catch((err) => (error = errorMessage(err)));
    editing = null;
    load();
  }
</script>

<div class="toolbar">
  <input class="input" type="search" placeholder="Filter…" bind:value={query} aria-label="Filter bookmarks" />
  <button class="btn btn-icon" type="button" aria-label="Add bookmark" onclick={() => (editing = 'new')}
    >＋</button
  >
</div>
{#if error}<p class="error">{error}</p>{/if}
{#if data}
  <ul class="list">
    {#each data.bookmarks as b (b.id)}
      {@const icon = faviconUrl(b.url)}
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
        {#each b.tags as t (t)}<span class="tag">#{t}</span>{/each}
        <button
          class="btn btn-icon edit"
          type="button"
          aria-label="Edit {b.title}"
          onclick={() => (editing = b)}>✎</button
        >
      </li>
    {:else}
      <li class="muted">No bookmarks{query ? ' match the filter' : ' yet'}.</li>
    {/each}
  </ul>
{/if}

{#if editing && data}
  <Modal title={editing === 'new' ? 'Add bookmark' : 'Edit bookmark'} onclose={() => (editing = null)}>
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
      <p><button class="btn btn-danger" type="button" onclick={() => remove(b)}>Delete bookmark</button></p>
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
