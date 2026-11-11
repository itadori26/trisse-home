<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';

  /** Props of the browser bookmark import. */
  interface Props {
    bookmarks: BookmarksUseCases;
  }
  let { bookmarks }: Props = $props();

  let message = $state('');
  let error = $state('');
  let busy = $state(false);

  async function onfile(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    busy = true;
    error = message = '';
    try {
      const r = await bookmarks.importHtml(await file.text());
      message = `${r.created} bookmarks imported (${r.foldersCreated} new folders). Skipped: ${r.skippedExists} already present, ${r.skippedInvalid} not importable.`;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<h4>{t('bookmarks.browser_bookmarks')}</h4>
<p class="muted">
  {t('bookmarks.export_your_bookmarks_as_html_in')}
</p>
<label class="btn file">
  {busy ? t('common.importing') : t('bookmarks.choose_file')}
  <input type="file" accept="text/html,.html,.htm" onchange={onfile} disabled={busy} />
</label>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  h4 {
    margin: var(--space-3) 0 var(--space-1);
  }
  .file input {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .file:focus-within {
    outline: 2px solid var(--color-accent);
  }
</style>
