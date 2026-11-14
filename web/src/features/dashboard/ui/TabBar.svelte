<script lang="ts">
  import { t } from '../../../shared/i18n';
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
    /** Opens the default-tab rules of a tab (#71). */
    onrules: (tab: Tab) => void;
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
    onrules,
    ontemplate,
  }: Props = $props();
</script>

<nav class="tabs" aria-label={t('dashboard.dashboard_tabs')} data-tour="tabs">
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
          aria-label={t('dashboard.move_tab_left')}
          disabled={i === 0}
          onclick={() => onmove(tab, -1)}>‹</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('dashboard.move_tab_right')}
          disabled={i === tabs.length - 1}
          onclick={() => onmove(tab, 1)}>›</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('dashboard.rename_tab')}
          onclick={() => onrename(tab)}>✎</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('dashboard.tab_background')}
          title={t('dashboard.background')}
          onclick={() => onbackground(tab)}>🖼</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('dashboard.tab_rules')}
          title={t('dashboard.tab_rules')}
          onclick={() => onrules(tab)}>⏱</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('dashboard.save_tab_as_template')}
          title={t('dashboard.save_as_template_2')}
          onclick={() => ontemplate(tab)}>⎘</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('dashboard.delete_tab')}
          disabled={tabs.length === 1}
          onclick={() => ondelete(tab)}>✕</button
        >
      {/if}
    </div>
  {/each}
  {#if editing}
    <button class="btn btn-icon" type="button" onclick={onadd}>{t('dashboard.tab')}</button>
  {/if}
</nav>

<style>
  /* Centred tab pills below the search (#189); overflow scrolls with faded edges. */
  .tabs {
    display: flex;
    flex-wrap: nowrap;
    justify-content: safe center;
    gap: var(--space-1);
    align-items: center;
    overflow-x: auto;
    scrollbar-width: none;
    max-width: 100%;
    padding: 2px 12px;
    mask-image: linear-gradient(to right, transparent 0, #000 12px, #000 calc(100% - 12px), transparent 100%);
  }
  .tabs::-webkit-scrollbar {
    display: none;
  }
  .tab {
    display: flex;
    align-items: center;
    flex: none;
    border-radius: 999px;
  }
  .name {
    border: none;
    background: none;
    padding: var(--space-1) var(--space-4);
    border-radius: 999px;
    cursor: pointer;
    color: var(--color-text-muted);
    font: inherit;
    font-weight: 500;
    white-space: nowrap;
  }
  .name:hover {
    color: var(--color-text);
  }
  .active .name {
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    font-weight: 600;
  }
  .name:focus-visible {
    outline: 2px solid var(--color-accent);
    outline-offset: 2px;
  }
</style>
