<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { BookmarkSettings } from '../domain/bookmark';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';

  /** Props of the opt-in title, favicon and link check settings (#72). */
  interface Props {
    bookmarks: BookmarksUseCases;
  }
  let { bookmarks }: Props = $props();

  let settings = $state<BookmarkSettings | null>(null);
  let error = $state('');
  let message = $state('');

  $effect(() => {
    bookmarks
      .getSettings()
      .then((s) => (settings = s))
      .catch((err) => (error = errorMessage(err)));
  });

  async function save() {
    if (!settings) return;
    error = message = '';
    try {
      settings = await bookmarks.saveSettings(settings);
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<h3>{t('bookmarks.links_heading')}</h3>
{#if settings}
  <label class="check">
    <input type="checkbox" bind:checked={settings.enrich} onchange={save} />
    {t('bookmarks.setting_enrich')}
  </label>
  <label class="check">
    <input type="checkbox" bind:checked={settings.checkLinks} onchange={save} />
    {t('bookmarks.setting_check_links')}
  </label>
  <p class="muted small">{t('bookmarks.links_privacy')}</p>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  .check {
    display: flex;
    gap: var(--space-2);
    align-items: flex-start;
    margin-bottom: var(--space-2);
  }
  .small {
    font-size: 0.85rem;
  }
</style>
