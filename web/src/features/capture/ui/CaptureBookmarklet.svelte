<script lang="ts">
  import { t } from '../../../shared/i18n';
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import { bookmarklet } from '../domain/capture';

  /** The bookmarklet of the quick capture page (#66), shown in Settings → Search. */
  const code = bookmarklet(location.origin);
</script>

<section class="bookmarklet">
  <h3>{t('capture.bookmarklet')}</h3>
  <p class="muted small">{t('capture.bookmarklet_hint')}</p>
  <!-- A javascript: link is the whole point of a bookmarklet; it only runs when the user clicks it in the bookmarks bar. -->
  <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
  <p>
    <a class="btn" href={code} onclick={(e) => e.preventDefault()} draggable="true"
      >{t('capture.bookmarklet_link')}</a
    >
  </p>
  <CopySnippet {code} label={t('capture.bookmarklet_copy')} />
</section>

<style>
  .bookmarklet {
    margin-top: var(--space-4, 1.5rem);
  }
  h3 {
    margin: 0 0 var(--space-1);
    font-size: 1em;
  }
  .small {
    font-size: 0.85em;
  }
</style>
