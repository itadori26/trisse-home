<script lang="ts">
  import { renderMarkdown } from './markdown';

  /** Props of the Markdown renderer. */
  interface Props {
    source: string;
  }
  let { source }: Props = $props();
  const html = $derived(renderMarkdown(source));
</script>

<!-- eslint-disable-next-line svelte/no-at-html-tags -- sanitised by DOMPurify in renderMarkdown -->
<div class="markdown">{@html html}</div>

<style>
  .markdown :global(:first-child) {
    margin-top: 0;
  }
  .markdown :global(pre),
  .markdown :global(code) {
    font-family: var(--font-mono);
    background: var(--color-surface-2);
    border-radius: 4px;
  }
  .markdown :global(pre) {
    padding: var(--space-2);
    overflow: auto;
  }
  .markdown :global(img) {
    max-width: 100%;
  }
  .markdown :global(ul),
  .markdown :global(ol) {
    padding-left: 1.4em;
  }
</style>
