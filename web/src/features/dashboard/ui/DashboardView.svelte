<script lang="ts">
  import { EditHistory, type EditCommand } from '../domain/history';
  import { t } from '../../../shared/i18n';
  import type { Snippet } from 'svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { IconName } from '../../../shared/ui/icons/paths';
  import WidgetPicker from './WidgetPicker.svelte';
  import type { Tab, Widget, WidgetConfig, WidgetType } from '../domain/dashboard';
  import type { LayoutItem } from '../domain/layout';
  import type { DashboardUseCases } from '../usecase/dashboardUseCases';
  import Grid from './Grid.svelte';
  import NewTabDialog from './NewTabDialog.svelte';
  import TabBar from './TabBar.svelte';
  import TextPrompt from './TextPrompt.svelte';
  import WidgetFrame from './WidgetFrame.svelte';

  /** Props of the dashboard page. */
  interface Props {
    uc: DashboardUseCases;
    /** Edit mode: drag, resize, add and remove widgets and tabs. */
    editing: boolean;
    /** Widget types that may be added (enabled modules only). */
    widgetTypes: WidgetType[];
    /** Picker icon per widget kind (#104). */
    widgetIcon: (kind: string) => IconName | undefined;
    /** Renders the content of a widget. */
    content: Snippet<[Widget]>;
    /** Renders the settings form of a widget; `save` persists a new config. */
    settings: Snippet<[Widget, (config: WidgetConfig) => Promise<void>]>;
    /** Whether a widget kind has a settings form. */
    hasSettings: (kind: string) => boolean;
    /** Whether a widget kind is rendered without card chrome outside edit mode. */
    chromeless: (kind: string) => boolean;
  }
  let { uc, editing, widgetTypes, widgetIcon, content, settings, hasSettings, chromeless }: Props = $props();

  const ACTIVE_TAB_KEY = 'trisse-home.activeTab';

  let tabs = $state<Tab[]>([]);
  let activeId = $state('');
  let error = $state('');
  let loaded = $state(false);
  let dialog = $state<
    | { kind: 'add-widget' }
    | { kind: 'new-tab' }
    | { kind: 'rename-tab'; tab: Tab }
    | { kind: 'background'; tab: Tab }
    | { kind: 'template'; tab: Tab }
    | { kind: 'settings'; widget: Widget }
    | null
  >(null);

  const active = $derived(tabs.find((t) => t.id === activeId) ?? tabs[0]);
  const titles = $derived(new Map(widgetTypes.map((t) => [t.kind, t.title])));
  const available = $derived(new Set(widgetTypes.map((t) => t.kind)));

  function readStoredTab(): string {
    try {
      return localStorage.getItem(ACTIVE_TAB_KEY) ?? '';
    } catch {
      return '';
    }
  }

  /** Selects the n-th tab (0-based), e.g. from the 1–9 shortcuts (#95). */
  export function selectTabAt(index: number) {
    const tab = tabs[index];
    if (tab) select(tab.id);
  }

  function select(id: string) {
    activeId = id;
    try {
      localStorage.setItem(ACTIVE_TAB_KEY, id);
    } catch {
      // storage may be unavailable (private mode); the selection is just not remembered
    }
  }

  async function run(action: () => Promise<void>) {
    error = '';
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function load() {
    await run(async () => {
      tabs = await uc.load();
      const stored = readStoredTab();
      activeId = tabs.some((t) => t.id === stored) ? stored : (tabs[0]?.id ?? '');
    });
    loaded = true;
  }

  function replaceTab(tab: Tab) {
    tabs = tabs.map((t) => (t.id === tab.id ? tab : t));
  }

  // Undo/redo of edit-mode operations within the session (#93).
  const history = new EditHistory();
  let historyVersion = $state(0);
  // The history is a plain object; historyVersion makes its state reactive.
  const canUndo = $derived.by(() => {
    void historyVersion;
    return history.canUndo;
  });
  const canRedo = $derived.by(() => {
    void historyVersion;
    return history.canRedo;
  });
  let toast = $state<{ text: string; key: number } | null>(null);
  let toastTimer: ReturnType<typeof setTimeout> | undefined;

  function record(c: EditCommand) {
    history.record(c);
    historyVersion++;
  }

  function showUndoToast(text: string) {
    clearTimeout(toastTimer);
    toast = { text, key: Date.now() };
    toastTimer = setTimeout(() => (toast = null), 8000);
  }

  const layoutOf = (widgets: readonly LayoutItem[]): LayoutItem[] =>
    widgets.map(({ id, x, y, w, h }) => ({ id, x, y, w, h }));

  function onlayout(items: LayoutItem[]) {
    const tab = active;
    if (!tab) return;
    const before = tab.widgets;
    const byId = new Map(items.map((i) => [i.id, i]));
    replaceTab({ ...tab, widgets: tab.widgets.map((w) => ({ ...w, ...byId.get(w.id) })) });
    record({ type: 'layout', tabId: tab.id, before: layoutOf(before), after: layoutOf(items) });
    run(() => uc.saveLayout(tab.id, before, items));
  }

  function addWidget(type: WidgetType) {
    const tab = active;
    if (!tab) return;
    dialog = null;
    run(async () => {
      const widget = await uc.addWidget(tab, type);
      replaceTab({ ...tab, widgets: [...tab.widgets, widget] });
      record({ type: 'add', widget });
      if (hasSettings(type.kind)) dialog = { kind: 'settings', widget };
    });
  }

  function duplicateWidget(widget: Widget) {
    const tab = tabs.find((t) => t.id === widget.tabId);
    if (!tab) return;
    run(async () => {
      const copy = await uc.duplicateWidget(tab, widget);
      replaceTab({ ...tab, widgets: [...tab.widgets, copy] });
      record({ type: 'add', widget: copy });
    });
  }

  const tabOf = (tabId: string) => tabs.find((t) => t.id === tabId);

  async function deleteById(tabId: string, id: string) {
    await uc.deleteWidget(id);
    const tab = tabOf(tabId);
    if (tab) replaceTab({ ...tab, widgets: tab.widgets.filter((w) => w.id !== id) });
  }

  async function recreate(widget: Widget) {
    const tab = tabOf(widget.tabId);
    if (!tab) throw new Error(t('dashboard.undo_tab_gone'));
    const created = await uc.restoreWidget(tab, { ...widget, id: history.resolve(widget.id) });
    history.remap(history.resolve(widget.id), created.id);
    replaceTab({ ...tab, widgets: [...tab.widgets, created] });
  }

  /** Applies a recorded command backwards (undo) or forwards (redo). */
  async function apply(c: EditCommand, undo: boolean) {
    switch (c.type) {
      case 'layout': {
        const tab = tabOf(c.tabId);
        if (!tab) return;
        const present = new Set(tab.widgets.map((w) => w.id));
        const items = history.resolveItems(undo ? c.before : c.after).filter((i) => present.has(i.id));
        const before = tab.widgets;
        const byId = new Map(items.map((i) => [i.id, i]));
        replaceTab({ ...tab, widgets: tab.widgets.map((w) => ({ ...w, ...byId.get(w.id) })) });
        await uc.saveLayout(tab.id, before, layoutOf(tabOf(c.tabId)?.widgets ?? []));
        return;
      }
      case 'add':
        if (undo) await deleteById(c.widget.tabId, history.resolve(c.widget.id));
        else await recreate(c.widget);
        return;
      case 'delete':
        if (undo) await recreate(c.widget);
        else await deleteById(c.widget.tabId, history.resolve(c.widget.id));
        return;
      case 'configure': {
        const id = history.resolve(c.widgetId);
        const updated = await uc.configureWidget(id, undo ? c.before : c.after);
        for (const tab of tabs) {
          if (tab.widgets.some((w) => w.id === id))
            replaceTab({
              ...tab,
              widgets: tab.widgets.map((w) => (w.id === id ? { ...w, config: updated.config } : w)),
            });
        }
      }
    }
  }

  /** Reverts the last edit-mode operation. */
  export function undo() {
    const c = history.takeUndo();
    if (!c) return;
    historyVersion++;
    toast = null;
    run(async () => {
      try {
        await apply(c, true);
      } catch (err) {
        history.restore(c, true);
        historyVersion++;
        throw err;
      }
    });
  }

  /** Applies the last undone operation again. */
  export function redo() {
    const c = history.takeRedo();
    if (!c) return;
    historyVersion++;
    run(async () => {
      try {
        await apply(c, false);
      } catch (err) {
        history.restore(c, false);
        historyVersion++;
        throw err;
      }
    });
  }

  // A tab background overrides the global one (see base.css).
  $effect(() => {
    const bg = active?.background ?? '';
    const root = document.documentElement.style;
    if (bg) root.setProperty('--tab-background-image', `url("${bg.replace(/["\\]/g, '')}")`);
    else root.removeProperty('--tab-background-image');
    return () => root.removeProperty('--tab-background-image');
  });

  // No confirmation any more: a deleted widget can be restored with Undo.
  function removeWidget(widget: Widget) {
    run(async () => {
      await deleteById(widget.tabId, widget.id);
      record({
        type: 'delete',
        widget: { ...widget, config: structuredClone($state.snapshot(widget.config)) },
      });
      showUndoToast(t('dashboard.widget_deleted', { name: titles.get(widget.kind) ?? widget.kind }));
    });
  }

  async function saveConfig(widget: Widget, config: WidgetConfig) {
    const before = structuredClone($state.snapshot(widget.config));
    const updated = await uc.configureWidget(widget.id, config);
    record({
      type: 'configure',
      widgetId: widget.id,
      before,
      after: structuredClone($state.snapshot(updated.config)),
    });
    const tab = tabs.find((t) => t.id === widget.tabId);
    if (tab)
      replaceTab({
        ...tab,
        widgets: tab.widgets.map((w) => (w.id === updated.id ? { ...w, config: updated.config } : w)),
      });
    dialog = null;
  }

  function deleteTab(tab: Tab) {
    if (!confirm(t('dashboard.confirm_delete_tab', { name: tab.name }))) return;
    run(async () => {
      await uc.deleteTab(tab.id);
      tabs = tabs.filter((t) => t.id !== tab.id);
      select(tabs[0]?.id ?? '');
    });
  }

  function moveTab(tab: Tab, delta: number) {
    const from = tabs.findIndex((t) => t.id === tab.id);
    run(async () => {
      tabs = await uc.moveTab(tabs, from, from + delta);
    });
  }

  load();
</script>

<section class="dashboard" data-testid="dashboard">
  <TabBar
    {tabs}
    activeId={active?.id ?? ''}
    {editing}
    onselect={select}
    onadd={() => (dialog = { kind: 'new-tab' })}
    onrename={(tab) => (dialog = { kind: 'rename-tab', tab })}
    ondelete={deleteTab}
    onmove={moveTab}
    onbackground={(tab) => (dialog = { kind: 'background', tab })}
    ontemplate={(tab) => (dialog = { kind: 'template', tab })}
  />

  {#if error}<p class="error" role="alert">{error}</p>{/if}

  {#if active}
    {@const visible = active.widgets.filter((w) => available.has(w.kind))}
    <Grid items={visible} {editing} {onlayout} label={(w) => titles.get(w.kind) ?? w.kind}>
      {#snippet item(widget: Widget, handle: { onpointerdown: (e: PointerEvent) => void })}
        <WidgetFrame
          title={titles.get(widget.kind) ?? widget.kind}
          {editing}
          {handle}
          chromeless={chromeless(widget.kind)}
          onsettings={hasSettings(widget.kind) ? () => (dialog = { kind: 'settings', widget }) : undefined}
          ondelete={() => removeWidget(widget)}
          onduplicate={() => duplicateWidget(widget)}
        >
          {@render content(widget)}
        </WidgetFrame>
      {/snippet}
    </Grid>
    {#if visible.length === 0 && loaded}
      <p class="empty muted">
        This tab is empty. {#if editing}Use “Add widget” to place one.{:else}Switch to edit mode to add
          widgets.{/if}
      </p>
    {/if}
    {#if editing}
      <div class="edit-actions">
        <button class="btn btn-primary add" type="button" onclick={() => (dialog = { kind: 'add-widget' })}
          >{t('dashboard.add_widget')}</button
        >
        <button class="btn" type="button" disabled={!canUndo} onclick={undo} aria-keyshortcuts="Control+Z"
          >↶ {t('dashboard.undo')}</button
        >
        <button
          class="btn"
          type="button"
          disabled={!canRedo}
          onclick={redo}
          aria-keyshortcuts="Control+Shift+Z">↷ {t('dashboard.redo')}</button
        >
      </div>
    {/if}
  {/if}
</section>

{#if toast}
  <div class="toast" role="status">
    {toast.text}
    <button class="btn" type="button" onclick={undo}>{t('dashboard.undo')}</button>
  </div>
{/if}

{#if dialog?.kind === 'add-widget'}
  <WidgetPicker types={widgetTypes} icon={widgetIcon} onpick={addWidget} onclose={() => (dialog = null)} />
{:else if dialog?.kind === 'new-tab'}
  <NewTabDialog
    {uc}
    onclose={() => (dialog = null)}
    oncreated={(tab) => {
      tabs = [...tabs, tab];
      select(tab.id);
    }}
  />
{:else if dialog?.kind === 'background'}
  {@const tab = dialog.tab}
  <TextPrompt
    title={t('dashboard.background_of', { name: tab.name })}
    label={t('dashboard.image_url_empty_global_background')}
    type="url"
    optional
    maxlength={2048}
    placeholder="https://…"
    initial={tab.background}
    onclose={() => (dialog = null)}
    onsubmit={async (url) => {
      const updated = await uc.setBackground(tab.id, url);
      replaceTab({ ...tab, background: updated.background });
    }}
  />
{:else if dialog?.kind === 'template'}
  {@const tab = dialog.tab}
  <TextPrompt
    title={t('dashboard.save_as_template', { name: tab.name })}
    label={t('dashboard.template_name')}
    initial={tab.name}
    onclose={() => (dialog = null)}
    onsubmit={async (name) => {
      await uc.saveTemplate(tab.id, name);
    }}
  />
{:else if dialog?.kind === 'rename-tab'}
  {@const tab = dialog.tab}
  <TextPrompt
    title={t('dashboard.rename_tab')}
    label={t('dashboard.name')}
    initial={tab.name}
    onclose={() => (dialog = null)}
    onsubmit={async (name) => {
      const renamed = await uc.renameTab(tab.id, name);
      replaceTab({ ...tab, name: renamed.name });
    }}
  />
{:else if dialog?.kind === 'settings'}
  {@const widget = dialog.widget}
  <Modal
    title={t('dashboard.settings', { value: titles.get(widget.kind) ?? widget.kind })}
    onclose={() => (dialog = null)}
  >
    {@render settings(widget, (config) => saveConfig(widget, config))}
  </Modal>
{/if}

<style>
  .dashboard {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .empty {
    text-align: center;
    padding: var(--space-6);
  }
  .add {
    align-self: center;
  }
  .edit-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: center;
    flex-wrap: wrap;
  }
  .toast {
    position: fixed;
    left: 50%;
    bottom: var(--space-4);
    transform: translateX(-50%);
    display: flex;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-small);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    box-shadow: 0 4px 16px rgb(0 0 0 / 20%);
    z-index: 20;
  }
</style>
