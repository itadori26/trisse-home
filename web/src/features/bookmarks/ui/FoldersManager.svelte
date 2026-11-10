<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Folder } from '../domain/bookmark';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';

  /** Props of the bookmark folder management section. */
  interface Props {
    bookmarks: BookmarksUseCases;
  }
  let { bookmarks }: Props = $props();

  let folders = $state<Folder[]>([]);
  let name = $state('');
  let error = $state('');

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      folders = (await bookmarks.list()).folders;
    } catch (err) {
      error = errorMessage(err);
    }
  }

  run(async () => {});
</script>

<p class="muted">Bookmarks are added and edited directly in the bookmarks widget.</p>
<ul class="list">
  {#each folders as f (f.id)}
    <li>
      <span class="name">{f.name} <span class="muted">({f.count})</span></span>
      <button
        class="btn btn-icon"
        type="button"
        aria-label="Rename {f.name}"
        onclick={() => {
          const n = prompt('Folder name', f.name);
          if (n) run(() => bookmarks.renameFolder(f.id, n));
        }}>✎</button
      >
      <button
        class="btn btn-icon"
        type="button"
        aria-label="Delete {f.name}"
        onclick={() =>
          confirm(`Delete folder "${f.name}"? Its bookmarks are kept.`) &&
          run(() => bookmarks.deleteFolder(f.id))}>✕</button
      >
    </li>
  {/each}
</ul>
<form
  class="add"
  onsubmit={(e) => {
    e.preventDefault();
    run(() => bookmarks.createFolder(name)).then(() => (name = ''));
  }}
>
  <input
    class="input"
    bind:value={name}
    placeholder="New folder"
    required
    maxlength="60"
    aria-label="Folder name"
  />
  <button class="btn" type="submit">Add</button>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    align-items: center;
  }
  .name {
    flex: 1;
  }
  .add {
    display: flex;
    gap: var(--space-2);
  }
</style>
