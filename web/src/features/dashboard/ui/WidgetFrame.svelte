<script lang="ts">
  import type { Snippet } from 'svelte';

  /** Props of the card that wraps every widget. */
  interface Props {
    title: string;
    editing: boolean;
    /** Spread onto the title bar so it acts as drag handle in edit mode. */
    handle: { onpointerdown: (e: PointerEvent) => void };
    /** Opens the widget settings; omitted for widgets without settings. */
    onsettings?: () => void;
    ondelete: () => void;
    /** Hides the title bar outside edit mode (e.g. for the search bar). */
    chromeless?: boolean;
    children: Snippet;
  }
  let { title, editing, handle, onsettings, ondelete, chromeless = false, children }: Props = $props();
</script>

<article class="widget card" class:chromeless={chromeless && !editing}>
  {#if editing || !chromeless}
    <!-- Dragging is pointer-only; settings and remove buttons stay keyboard accessible. -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <header class:draggable={editing} onpointerdown={handle.onpointerdown}>
      <h3>{title}</h3>
      {#if editing}
        <span class="actions">
          {#if onsettings}
            <button
              class="btn btn-icon"
              type="button"
              aria-label="Widget settings"
              onpointerdown={(e) => e.stopPropagation()}
              onclick={onsettings}>⚙</button
            >
          {/if}
          <button
            class="btn btn-icon"
            type="button"
            aria-label="Remove widget"
            onpointerdown={(e) => e.stopPropagation()}
            onclick={ondelete}>✕</button
          >
        </span>
      {/if}
    </header>
  {/if}
  <div class="content">{@render children()}</div>
</article>

<style>
  .widget {
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    opacity: var(--widget-opacity);
  }
  .widget.chromeless {
    background: transparent;
    border-color: transparent;
    box-shadow: none;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-2) var(--space-3) 0;
    min-height: 32px;
  }
  header.draggable {
    cursor: grab;
  }
  h3 {
    margin: 0;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .content {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: var(--space-2) var(--space-3) var(--space-3);
  }
  .chromeless .content {
    padding: 0;
    overflow: visible;
  }
</style>
