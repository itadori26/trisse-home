<script lang="ts">
  import { renderMarkdown } from './markdown';

  /** Props of the Markdown renderer. */
  interface Props {
    source: string;
    /** Receives clicks on in-page links (href "#…"), e.g. links to other notes. */
    onlink?: (href: string) => void;
  }
  let { source, onlink }: Props = $props();
  const html = $derived(renderMarkdown(source));

  function click(e: MouseEvent) {
    const a = (e.target as Element | null)?.closest('a');
    const href = a?.getAttribute('href') ?? '';
    if (onlink && href.startsWith('#')) {
      e.preventDefault();
      onlink(href);
    }
  }
</script>

<!-- Links are focusable and activated by keyboard; the handler only intercepts their default. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<!-- eslint-disable-next-line svelte/no-at-html-tags -- sanitised by DOMPurify in renderMarkdown -->
<div class="markdown" onclick={click}>{@html html}</div>

<style>
  .markdown :global(:first-child) {
    margin-top: 0;
  }
  /* Notes live in small widgets: keep headings modest. */
  .markdown :global(h1) {
    font-size: 1.3rem;
  }
  .markdown :global(h2) {
    font-size: 1.15rem;
  }
  .markdown :global(h3) {
    font-size: 1rem;
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
