<script lang="ts">
  import type { Tab } from '../domain/dashboard';

  /** Props of the tab bar. */
  interface Props {
    tabs: Tab[];
    activeId: string;
    editing: boolean;
    onselect: (id: string) => void;
    onadd: () => void;
    onrename: (tab: Tab) => void;
    ondelete: (tab: Tab) => void;
    /** Moves a tab by `delta` positions (-1 = left, +1 = right). */
    onmove: (tab: Tab, delta: number) => void;
    /** Opens the background dialog of a tab. */
    onbackground: (tab: Tab) => void;
    /** Saves a tab as template. */
    ontemplate: (tab: Tab) => void;
  }
  let {
    tabs,
    activeId,
    editing,
    onselect,
    onadd,
    onrename,
    ondelete,
    onmove,
    onbackground,
    ontemplate,
  }: Props = $props();
</script>

<nav class="tabs" aria-label="Dashboard tabs" data-tour="tabs">
  {#each tabs as tab, i (tab.id)}
    <div class="tab" class:active={tab.id === activeId}>
      <button
        type="button"
        class="name"
        data-tour="tab-{tab.name}"
        aria-current={tab.id === activeId ? 'page' : undefined}
        onclick={() => onselect(tab.id)}
      >
        {tab.name}
      </button>
      {#if editing && tab.id === activeId}
        <button
          class="btn btn-icon"
          type="button"
          aria-label="Move tab left"
          disabled={i === 0}
          onclick={() => onmove(tab, -1)}>‹</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label="Move tab right"
          disabled={i === tabs.length - 1}
          onclick={() => onmove(tab, 1)}>›</button
        >
        <button class="btn btn-icon" type="button" aria-label="Rename tab" onclick={() => onrename(tab)}
          >✎</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label="Tab background"
          title="Background"
          onclick={() => onbackground(tab)}>🖼</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label="Save tab as template"
          title="Save as template"
          onclick={() => ontemplate(tab)}>⎘</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label="Delete tab"
          disabled={tabs.length === 1}
          onclick={() => ondelete(tab)}>✕</button
        >
      {/if}
    </div>
  {/each}
  {#if editing}
    <button class="btn btn-icon" type="button" onclick={onadd}>+ Tab</button>
  {/if}
</nav>

<style>
  .tabs {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    align-items: center;
  }
  .tab {
    display: flex;
    align-items: center;
    border-radius: 999px;
  }
  .tab.active {
    background: var(--color-surface);
    box-shadow: var(--shadow);
  }
  .name {
    border: none;
    background: none;
    padding: var(--space-2) var(--space-4);
    cursor: pointer;
    color: var(--color-text-muted);
    font-weight: 500;
  }
  .active .name {
    color: var(--color-text);
  }
</style>
