<script lang="ts">
  import { parseBookmarksConfig, type Folder, type Tag, UNFILED } from '../domain/bookmark';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';

  /** Props of the bookmarks widget settings. */
  interface Props {
    bookmarks: BookmarksUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { bookmarks, config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state(parseBookmarksConfig(config));
  let folders = $state<Folder[]>([]);
  let tags = $state<Tag[]>([]);

  $effect(() => {
    bookmarks.list().then((l) => ((folders = l.folders), (tags = l.tags)));
  });
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field">
    <span>Folder</span>
    <select class="input" bind:value={form.folderId}>
      <option value="">All bookmarks</option>
      <option value={UNFILED}>Without folder</option>
      {#each folders as f (f.id)}<option value={f.id}>{f.name}</option>{/each}
    </select>
  </label>
  <label class="field">
    <span>Tag</span>
    <select class="input" bind:value={form.tag}>
      <option value="">Any tag</option>
      {#each tags as t (t.name)}<option value={t.name}>#{t.name} ({t.count})</option>{/each}
    </select>
  </label>
  <button class="btn btn-primary" type="submit">Save</button>
</form>
