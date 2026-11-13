<script lang="ts">
  import { t } from '../../../shared/i18n';
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
    <span>{t('bookmarks.folder')}</span>
    <select class="input" bind:value={form.folderId}>
      <option value="">{t('bookmarks.all_bookmarks')}</option>
      <option value={UNFILED}>{t('bookmarks.without_folder')}</option>
      {#each folders as f (f.id)}<option value={f.id}>{f.name}</option>{/each}
    </select>
  </label>
  <label class="field">
    <span>{t('bookmarks.tag')}</span>
    <select class="input" bind:value={form.tag}>
      <option value="">{t('bookmarks.any_tag')}</option>
      {#each tags as t (t.name)}<option value={t.name}>#{t.name} ({t.count})</option>{/each}
    </select>
  </label>
  <label class="field">
    <span>{t('bookmarks.sort')}</span>
    <select class="input" bind:value={form.sort}>
      <option value="manual">{t('bookmarks.sort_manual')}</option>
      <option value="az">{t('bookmarks.sort_az')}</option>
      <option value="newest">{t('bookmarks.sort_newest')}</option>
    </select>
  </label>
  <button class="btn btn-primary" type="submit">{t('bookmarks.save')}</button>
</form>
