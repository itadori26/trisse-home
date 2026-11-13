<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Bookmark, type Folder, linkProblem, parseTags } from '../domain/bookmark';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';

  /** Props of the bookmark create/edit form. */
  interface Props {
    bookmarks: BookmarksUseCases;
    folders: Folder[];
    /** Bookmark to edit; omitted to create a new one. */
    bookmark?: Bookmark;
    /** Folder preselected for new bookmarks. */
    folderId?: string;
    onsaved: () => void;
  }
  let { bookmarks, folders, bookmark, folderId = '', onsaved }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state({
    title: bookmark?.title ?? '',
    url: bookmark?.url ?? '',
    description: bookmark?.description ?? '',
    tags: bookmark?.tags.join(', ') ?? '',
    folderId: bookmark?.folderId ?? folderId,
  });
  let error = $state('');
  let info = $state('');
  let busy = $state(false);

  /** Re-fetches title and icon or checks the link of a saved bookmark (#72). */
  async function action(kind: 'enrich' | 'check') {
    if (!bookmark) return;
    error = info = '';
    busy = true;
    try {
      const b =
        kind === 'enrich' ? await bookmarks.enrich(bookmark.id) : await bookmarks.checkLink(bookmark.id);
      if (kind === 'enrich') form.title = b.title;
      info =
        kind === 'enrich'
          ? b.hasIcon
            ? t('bookmarks.enriched')
            : t('bookmarks.enriched_no_icon')
          : linkProblem(b.link, intlLocale()) || t('bookmarks.link_ok');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      await bookmarks.save({ ...form, tags: parseTags(form.tags) }, bookmark?.id);
      onsaved();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field"
    ><span>URL</span><input class="input" bind:value={form.url} required placeholder="https://…" /></label
  >
  <label class="field"
    ><span>{t('bookmarks.title_optional')}</span><input
      class="input"
      bind:value={form.title}
      maxlength="200"
    /></label
  >
  <label class="field">
    <span>{t('bookmarks.folder')}</span>
    <select class="input" bind:value={form.folderId}>
      <option value="">{t('bookmarks.none')}</option>
      {#each folders as f (f.id)}<option value={f.id}>{f.name}</option>{/each}
    </select>
  </label>
  <label class="field"
    ><span>{t('bookmarks.tags_comma_separated')}</span><input
      class="input"
      bind:value={form.tags}
      placeholder={t('bookmarks.dev_docs')}
    /></label
  >
  <label class="field"
    ><span>{t('bookmarks.description')}</span><input
      class="input"
      bind:value={form.description}
      maxlength="1000"
    /></label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if info}<p class="muted" role="status">{info}</p>{/if}
  <p class="actions">
    <button class="btn btn-primary" type="submit">{t('bookmarks.save')}</button>
    {#if bookmark}
      <button class="btn" type="button" disabled={busy} onclick={() => action('enrich')}
        >{t('bookmarks.fetch_title_icon')}</button
      >
      <button class="btn" type="button" disabled={busy} onclick={() => action('check')}
        >{t('bookmarks.check_link')}</button
      >
    {/if}
  </p>
</form>

<style>
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
</style>
