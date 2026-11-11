<script lang="ts">
  import { t } from '../i18n';

  /** Props of a copy-ready code snippet. */
  interface Props {
    code: string;
    label: string;
  }
  let { code, label }: Props = $props();

  let copied = $state(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(code);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // Clipboard API unavailable (e.g. plain HTTP): the text stays selectable.
      copied = false;
    }
  }
</script>

<figure class="snippet">
  <pre aria-label={label}><code>{code}</code></pre>
  <button class="btn" type="button" onclick={copy} aria-live="polite">
    {copied ? t('common.copied') : t('common.copy')}
  </button>
</figure>

<style>
  .snippet {
    position: relative;
    margin: var(--space-2) 0;
  }
  pre {
    margin: 0;
    padding: var(--space-2) var(--space-3);
    padding-right: 6rem;
    overflow-x: auto;
    border-radius: var(--radius-small);
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
    font-size: 0.85em;
    user-select: all;
  }
  .btn {
    position: absolute;
    top: var(--space-1);
    right: var(--space-1);
  }
</style>
