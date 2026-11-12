<script lang="ts">
  import { t } from '../i18n';
  import type { Snippet } from 'svelte';

  /** Props of the modal dialog. */
  interface Props {
    /** Heading shown at the top of the dialog. */
    title: string;
    /** Called when the dialog is dismissed (Escape, backdrop, close button). */
    onclose: () => void;
    children: Snippet;
    /** "wide" is a landscape dialog that becomes full-screen on phones. */
    size?: 'normal' | 'wide';
  }

  let { title, onclose, children, size = 'normal' }: Props = $props();
  let dialog: HTMLDialogElement;

  $effect(() => {
    dialog.showModal();
    return () => dialog.close();
  });
</script>

<dialog
  bind:this={dialog}
  class="card"
  class:wide={size === 'wide'}
  aria-label={title}
  oncancel={(e) => {
    e.preventDefault();
    onclose();
  }}
  onclick={(e) => {
    if (e.target === dialog) onclose();
  }}
>
  <header>
    <h2>{title}</h2>
    <button class="btn btn-icon" type="button" aria-label={t('ui.close')} onclick={onclose}>✕</button>
  </header>
  <div class="body">{@render children()}</div>
</dialog>

<style>
  dialog {
    width: min(560px, calc(100vw - 32px));
    max-height: calc(100vh - 64px);
    padding: 0;
    color: var(--color-text);
  }
  dialog.wide {
    width: min(1100px, 95vw);
  }
  @media (max-width: 640px) {
    dialog.wide {
      width: 100vw;
      max-width: 100vw;
      height: 100dvh;
      max-height: 100dvh;
      margin: 0;
      border-radius: 0;
    }
  }
  dialog::backdrop {
    background: rgb(0 0 0 / 45%);
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }
  h2 {
    margin: 0;
    font-size: 1.1rem;
  }
  .body {
    padding: var(--space-4);
    overflow: auto;
  }
</style>
