<script lang="ts">
  import { EditHistory, type EditCommand, layoutCommand } from '../domain/history';
  import { t } from '../../../shared/i18n';
  import { backgroundCss } from '../../../shared/theme/background';
  import type { Snippet } from 'svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { IconName } from '../../../shared/ui/icons/paths';
  import WidgetPicker from './WidgetPicker.svelte';
  import {
    initialTabId,
    type Tab,
    type Widget,
    type WidgetConfig,
    type WidgetType,
  } from '../domain/dashboard';
  import { type LayoutItem, layoutCells, setItemSize } from '../domain/layout';
  import { isCramped, type WidgetSize, widgetView } from '../domain/sizes';
  import { SvelteSet } from 'svelte/reactivity';
  import DetailPanel from './DetailPanel.svelte';
  import type { DashboardUseCases } from '../usecase/dashboardUseCases';
  import Grid from './Grid.svelte';
  import MobileList from './MobileList.svelte';
  import { deriveMobile, type MobileItem, PHONE_QUERY, resolveMobile } from '../domain/mobile';
  import NewTabDialog from './NewTabDialog.svelte';
  import TabBar from './TabBar.svelte';
  import { portal } from '../../../shared/ui/portal';
  import { showTabBar } from '../domain/header';
  import { tick } from 'svelte';
  import { highlightWidget, onWidgetFocusRequest, pickWidget } from '../../../shared/ui/widgetFocus';
  import { announce } from '../../../shared/a11y/live.svelte';
  import TabRuleOffer from './TabRuleOffer.svelte';
  import TabRulesDialog from './TabRulesDialog.svelte';
  import TextPrompt from './TextPrompt.svelte';
  import WidgetFrame from './WidgetFrame.svelte';
  import ZoneLayout from './ZoneLayout.svelte';
  import { groupByZone, type Zone, zoneOf } from '../domain/zones';

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
    /** Id of an element in the shell's top bar that hosts the tab bar (#167). */
    tabBarTarget?: string;
    /** Reports the loaded tabs and the active one, e.g. for the greeting line. */
    onstate?: (tabs: Tab[], activeId: string) => void;
    /** Starts edit mode (the "Add widget" button of an empty tab, #276). */
    onedit?: () => void;
    /** Finishes edit mode (the Done button of the action bar, #271). */
    ondone?: () => void;
    /**
     * Opens the background picker for a tab (#277). It lives outside this
     * feature because it also changes the global background; without it
     * the tab background is a plain URL prompt.
     */
    onbackground?: (tab: Tab) => void;
    /** Extra content below the widget picker, e.g. what a server adds (#291). */
    pickerFooter?: Snippet;
    /** Renders the compact view of a widget (#595); strip is a one-line tile. */
    compact?: Snippet<[Widget, { strip: boolean }]>;
    /** Whether a widget has a compact view with its current configuration. */
    hasCompact?: (widget: Widget) => boolean;
    /** Whether compact tiles open the detail panel; off on kiosk displays (#594). */
    detailPanel?: boolean;
    /**
     * "Notify me when this fails" in the detail panel (#594): what it does
     * for a widget, or undefined where it is not offered (no admin, a widget
     * that watches nothing).
     */
    notifyAction?: (widget: Widget) => (() => void) | undefined;
    /** Example of a kind's compact view for the size previews of the picker (#596). */
    compactExample?: (kind: string) => string | undefined;
    /** Whether a widget watches the system (services, backups, drives; #593). */
    statusWidget?: (widget: Widget) => boolean;
    /**
     * Lines above the grid (#593): the attention line and the system
     * summary. `status` are the tab's status widgets, `collapsed` whether
     * they are folded into the summary, `view` shows a widget.
     */
    overview?: Snippet<[Overview]>;
    /** Whether the viewer sees the System zone of zoned tabs; admins only (#619). */
    systemZone?: boolean;
    /** Whether a widget has nothing to show; zoned tabs hide it outside edit mode (#619). */
    emptyWidget?: (widget: Widget) => boolean;
  }

  /** What the overview lines above the grid get. */
  interface Overview {
    tab: Tab;
    status: Widget[];
    collapsed: boolean;
    toggle: () => void;
    view: (widgetId: string) => void;
  }
  let {
    uc,
    editing,
    widgetTypes,
    widgetIcon,
    content,
    settings,
    hasSettings,
    chromeless,
    tabBarTarget = '',
    onstate,
    onedit,
    ondone,
    onbackground,
    pickerFooter,
    compact,
    hasCompact = () => false,
    detailPanel = true,
    notifyAction = () => undefined,
    compactExample,
    statusWidget = () => false,
    overview,
    systemZone = true,
    emptyWidget = () => false,
  }: Props = $props();

  // Tabs whose status widgets are folded into the system summary (#593).
  // A per-browser view choice, like the active tab, so it lives in storage.
  const COLLAPSED_KEY = 'waylight.collapsedStatus';
  const collapsedTabs = new SvelteSet<string>(readCollapsed());
  function readCollapsed(): string[] {
    try {
      const ids: unknown = JSON.parse(localStorage.getItem(COLLAPSED_KEY) ?? '[]');
      return Array.isArray(ids) ? ids.filter((id): id is string => typeof id === 'string') : [];
    } catch {
      return [];
    }
  }
  function toggleCollapsed(tabId: string) {
    if (collapsedTabs.has(tabId)) collapsedTabs.delete(tabId);
    else collapsedTabs.add(tabId);
    try {
      localStorage.setItem(COLLAPSED_KEY, JSON.stringify([...collapsedTabs]));
    } catch {
      // storage may be unavailable (private mode); the choice is just not remembered
    }
  }

  /** Shows a widget from the overview: unfolds the summary first if it hides it. */
  async function viewWidget(tab: Tab, widgetId: string) {
    const widget = tab.widgets.find((w) => w.id === widgetId);
    if (widget && statusWidget(widget) && collapsedTabs.has(tab.id)) toggleCollapsed(tab.id);
    await tick();
    highlightWidget(widgetId, true);
  }

  // The widget whose full view the detail panel shows, and the tile that opened it.
  let detail = $state<{ widget: Widget; opener: HTMLElement } | null>(null);
  // The panel belongs to the tab it was opened on; it closes when the tab
  // changes or the widget is gone (deleted, or undone in another step).
  $effect(() => {
    const open = detail;
    if (open && !active?.widgets.some((w) => w.id === open.widget.id)) detail = null;
  });

  /** "Edit widget" in the panel: the settings form, for whoever may edit the tab. */
  function editFromPanel(widget: Widget) {
    detail = null;
    dialog = { kind: 'settings', widget };
  }

  // Tiles whose content box is too small for the full view right now (#595).
  const cramped = new SvelteSet<string>();
  function onbox(id: string, width: number, height: number) {
    if (isCramped(width, height)) cramped.add(id);
    else cramped.delete(id);
  }

  const ACTIVE_TAB_KEY = 'waylight.activeTab';
  // A tab picked by hand wins over rules until the browser is closed (#71).
  const MANUAL_TAB_KEY = 'waylight.manualTab';

  let tabs = $state<Tab[]>([]);
  let activeId = $state('');
  let error = $state('');
  let loaded = $state(false);
  // The client's private network for an "in this network" rule (#673).
  let clientNetwork = $state('');
  // The tab just created, which is offered rule presets once (#673). Kept in
  // memory only: after a reload the offer is gone for good.
  let offerTabId = $state('');
  let dialog = $state<
    | { kind: 'add-widget'; query?: string }
    | { kind: 'new-tab' }
    | { kind: 'rename-tab'; tab: Tab }
    | { kind: 'delete-tab'; tab: Tab }
    | { kind: 'background'; tab: Tab }
    | { kind: 'rules'; tab: Tab }
    | { kind: 'template'; tab: Tab }
    | { kind: 'settings'; widget: Widget }
    | null
  >(null);

  const active = $derived(tabs.find((t) => t.id === activeId) ?? tabs[0]);
  // A team tab (#295) stays read-only in edit mode; the bar still lets the
  // viewer arrange their own tabs.
  const editingTab = $derived(editing && (active?.mine ?? true));

  $effect(() => onstate?.(tabs, active?.id ?? ''));

  // Phones get their own single-column layout per tab (#92).
  const phoneQuery = typeof matchMedia === 'function' ? matchMedia(PHONE_QUERY) : null;
  let phone = $state(phoneQuery?.matches ?? false);
  $effect(() => {
    if (!phoneQuery) return;
    const update = () => (phone = phoneQuery.matches);
    phoneQuery.addEventListener('change', update);
    return () => phoneQuery.removeEventListener('change', update);
  });

  // Other widgets can ask to show a widget, e.g. the briefing (#52).
  $effect(() =>
    onWidgetFocusRequest(async (kinds) => {
      const target = pickWidget(tabs, active?.id ?? '', kinds);
      if (!target) return;
      if (target.tabId !== active?.id) select(target.tabId);
      await tick();
      highlightWidget(target.widgetId);
    }),
  );

  // A new widget is shown, outlined and focused; with a settings form only
  // after that dialog closed (#273).
  let pendingHighlight = $state<string | null>(null);
  $effect(() => {
    const id = pendingHighlight;
    if (!id || dialog) return;
    pendingHighlight = null;
    void tick().then(() => highlightWidget(id, true));
  });

  // Entering edit mode is announced once (#271).
  let wasEditing = false;
  $effect(() => {
    if (editing && !wasEditing) announce(t('dashboard.edit_mode_hint'));
    wasEditing = editing;
  });

  function onmobile(tab: Tab, layout: MobileItem[]) {
    replaceTab({ ...tab, mobile: layout, mobileCustom: true });
    run(async () => {
      const saved = await uc.saveMobileLayout(tab.id, layout);
      const current = tabOf(tab.id);
      if (current) replaceTab({ ...current, mobile: saved.mobile, mobileCustom: saved.mobileCustom });
    });
  }

  function resetMobile(tab: Tab) {
    run(async () => {
      const saved = await uc.resetMobileLayout(tab.id);
      const current = tabOf(tab.id);
      if (current) replaceTab({ ...current, mobile: saved.mobile, mobileCustom: false });
    });
  }
  const titles = $derived(new Map(widgetTypes.map((t) => [t.kind, t.title])));
  const categories = $derived(new Map(widgetTypes.map((t) => [t.kind, t.category])));
  const categoryOf = (kind: string) => categories.get(kind);

  /** Switches the zone layout of a tab on or off (#619). */
  function toggleZones(tab: Tab) {
    run(async () => {
      const saved = await uc.setZones(tab.id, !tab.zoned);
      const current = tabOf(tab.id);
      if (current) replaceTab({ ...current, zoned: saved.zoned });
    });
  }

  /** Moves a widget to another zone of its tab (#619). */
  function moveToZone(widget: Widget, zone: Zone) {
    const tab = tabOf(widget.tabId);
    if (!tab) return;
    replaceTab({ ...tab, widgets: tab.widgets.map((w) => (w.id === widget.id ? { ...w, zone } : w)) });
    run(() => uc.moveToZone(widget, zone));
  }
  const available = $derived(new Set(widgetTypes.map((t) => t.kind)));

  function readStored(storage: () => Storage, key: string): string {
    try {
      return storage().getItem(key) ?? '';
    } catch {
      return '';
    }
  }

  /** Opens the widget picker, e.g. from "add widget weather" in the search (#476). */
  export function openAddWidget(query = '') {
    dialog = { kind: 'add-widget', query };
  }

  /** Selects a tab by id, e.g. from "tab arbeit" in the search (#476). */
  export function selectTab(id: string) {
    if (tabs.some((tab) => tab.id === id)) select(id);
  }

  /**
   * Selects the widget's tab and opens its settings, e.g. from the
   * needs-your-attention list (#592); does nothing for an unknown widget.
   */
  export function openWidgetSettings(tabId: string, widgetId: string) {
    const widget = tabs.find((tab) => tab.id === tabId)?.widgets.find((w) => w.id === widgetId);
    if (!widget) return;
    select(tabId);
    if (hasSettings(widget.kind)) dialog = { kind: 'settings', widget };
  }

  /** Shows a tab created outside the view, e.g. from a template on first start (#589). */
  export function addTab(tab: Tab) {
    tabs = [...tabs, tab];
    select(tab.id);
  }

  /**
   * Loads the tabs again, e.g. after the template choice on first start
   * replaced the example tabs (#670), and selects the given tab.
   */
  export async function reload(selectId?: string) {
    await load();
    if (selectId) selectTab(selectId);
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
      sessionStorage.setItem(MANUAL_TAB_KEY, id);
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

  // A link with ?tab=<id> opens that tab once (#568), e.g. from the mobile
  // app; the parameter is dropped so a reload follows the usual choice.
  function requestedTab(): string {
    const url = new URL(location.href);
    const id = url.searchParams.get('tab') ?? '';
    if (id) {
      url.searchParams.delete('tab');
      // window.history: `history` here is the edit history of the tab.
      window.history.replaceState(window.history.state, '', url);
    }
    return id;
  }

  async function load() {
    const requested = requestedTab();
    await run(async () => {
      const dashboard = await uc.open();
      tabs = dashboard.tabs;
      clientNetwork = dashboard.clientNetwork ?? '';
      activeId = initialTabId(
        tabs.map((t) => t.id),
        {
          requested,
          manual: readStored(() => sessionStorage, MANUAL_TAB_KEY),
          byRule: dashboard.defaultTabId,
          lastUsed: readStored(() => localStorage, ACTIVE_TAB_KEY),
        },
      );
    });
    loaded = true;
  }

  function openBackground(tab: Tab) {
    if (onbackground) onbackground(tab);
    else dialog = { kind: 'background', tab };
  }

  /** Shows a tab background saved by the picker outside this view (#277). */
  export function showTabBackground(tabId: string, background: string) {
    const tab = tabs.find((x) => x.id === tabId);
    if (tab) replaceTab({ ...tab, background });
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

  /** A widget at the place and size of a layout item; no size makes it free. */
  function place(widget: Widget, item: LayoutItem | undefined): Widget {
    if (!item) return widget;
    const next: Widget = { ...widget, x: item.x, y: item.y, w: item.w, h: item.h };
    if (item.size) next.size = item.size;
    else delete next.size;
    return next;
  }

  function onlayout(items: LayoutItem[]) {
    const tab = active;
    if (!tab) return;
    const before = tab.widgets;
    const command = layoutCommand(tab.id, before, items);
    if (!command) return; // nothing moved: no undo step, no save (#272)
    const byId = new Map(items.map((i) => [i.id, i]));
    replaceTab({ ...tab, widgets: tab.widgets.map((w) => place(w, byId.get(w.id))) });
    record(command);
    run(() => uc.saveLayout(tab.id, before, items));
  }

  /** Gives a widget one of the sizes or makes it freely sized again (#595). */
  function setSize(widget: Widget, size: WidgetSize | undefined) {
    const tab = tabOf(widget.tabId);
    if (tab) onlayout(setItemSize(layoutCells(tab.widgets), widget.id, size));
  }

  function addWidget(type: WidgetType, size?: WidgetSize) {
    const tab = active;
    if (!tab) return;
    dialog = null;
    run(async () => {
      const widget = await uc.addWidget(tab, type, undefined, size);
      replaceTab({ ...tab, widgets: [...tab.widgets, widget] });
      record({ type: 'add', widget });
      if (hasSettings(type.kind)) dialog = { kind: 'settings', widget };
      pendingHighlight = widget.id;
    });
  }

  function duplicateWidget(widget: Widget) {
    const tab = tabs.find((t) => t.id === widget.tabId);
    if (!tab) return;
    run(async () => {
      const copy = await uc.duplicateWidget(tab, widget);
      replaceTab({ ...tab, widgets: [...tab.widgets, copy] });
      record({ type: 'add', widget: copy });
      pendingHighlight = copy.id;
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
        replaceTab({ ...tab, widgets: tab.widgets.map((w) => place(w, byId.get(w.id))) });
        await uc.saveLayout(tab.id, before, layoutCells(tabOf(c.tabId)?.widgets ?? []));
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
    const bg = backgroundCss(active?.background ?? '');
    const root = document.documentElement.style;
    if (bg) root.setProperty('--tab-background-image', bg);
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

  /** Asks in the app's own dialog instead of window.confirm() (#274). */
  function deleteTab(tab: Tab) {
    if (tabs.length > 1) dialog = { kind: 'delete-tab', tab };
  }

  function confirmDeleteTab(tab: Tab) {
    dialog = null;
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
  {#if showTabBar(tabs.length, editing)}
    <!-- Rendered into the shell's top bar when it offers a place (#167). -->
    <div class="tab-host" use:portal={tabBarTarget}>
      <TabBar
        {tabs}
        activeId={active?.id ?? ''}
        {editing}
        onselect={select}
        onadd={() => (dialog = { kind: 'new-tab' })}
        onrename={(tab) => (dialog = { kind: 'rename-tab', tab })}
        ondelete={deleteTab}
        onmove={moveTab}
        onbackground={(tab) => openBackground(tab)}
        onrules={(tab) => (dialog = { kind: 'rules', tab })}
        ontemplate={(tab) => (dialog = { kind: 'template', tab })}
      />
    </div>
  {/if}

  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if editing && !editingTab}<p class="edit-hint">{t('dashboard.team_tab_read_only')}</p>
  {:else if editing}<p class="edit-hint">{t('dashboard.edit_mode_hint')}</p>{/if}
  {#if active && active.id === offerTabId}
    {@const tab = active}
    <TabRuleOffer
      {tab}
      {clientNetwork}
      onsubmit={async (rules) => {
        const updated = await uc.setRules(tab.id, rules);
        replaceTab({ ...tab, rules: updated.rules });
      }}
      onedit={() => {
        offerTabId = '';
        dialog = { kind: 'rules', tab };
      }}
      onclose={() => (offerTabId = '')}
    />
  {/if}

  {#if active}
    {@const status = active.widgets.filter((w) => available.has(w.kind) && statusWidget(w))}
    {@const folded = collapsedTabs.has(active.id) && !editing && status.length > 0}
    {@const zoned = active.zoned}
    {@const visible = active.widgets.filter(
      (w) =>
        available.has(w.kind) &&
        !(folded && statusWidget(w)) &&
        // The quiet look of zones (#619): a widget with nothing to show is left out.
        !(zoned && !editing && emptyWidget(w)),
    )}
    {#if overview && !editing}
      {@const tab = active}
      {@render overview({
        tab,
        status,
        collapsed: folded,
        toggle: () => toggleCollapsed(tab.id),
        view: (id) => viewWidget(tab, id),
      })}
    {/if}
    {#snippet frame(
      widget: Widget,
      handle: { onpointerdown: (e: PointerEvent) => void },
      size: WidgetSize | undefined,
      sizable: boolean,
    )}
      {@const view = compact
        ? widgetView({ size, hasCompact: hasCompact(widget), cramped: cramped.has(widget.id) })
        : 'full'}
      <WidgetFrame
        title={titles.get(widget.kind) ?? widget.kind}
        editing={editingTab}
        {handle}
        chromeless={chromeless(widget.kind)}
        compact={view === 'compact'}
        onopen={view === 'compact' && detailPanel ? (opener) => (detail = { widget, opener }) : undefined}
        size={widget.size}
        onsize={sizable ? (s) => setSize(widget, s) : undefined}
        zone={zoned ? zoneOf(widget, categoryOf(widget.kind)) : undefined}
        onzone={zoned ? (z) => moveToZone(widget, z) : undefined}
        onbox={(w, h) => onbox(widget.id, w, h)}
        onsettings={hasSettings(widget.kind) ? () => (dialog = { kind: 'settings', widget }) : undefined}
        ondelete={() => removeWidget(widget)}
        onduplicate={() => duplicateWidget(widget)}
      >
        {#if view === 'compact' && compact}
          {@render compact(widget, { strip: size === 'strip' })}
        {:else}
          {@render content(widget)}
        {/if}
      </WidgetFrame>
    {/snippet}
    {#if zoned}
      {@const tab = active}
      {@const phoneLayout = phone
        ? tab.mobileCustom
          ? resolveMobile(tab.mobile, visible)
          : deriveMobile(visible)
        : []}
      {@const phoneSizes = new Map(phoneLayout.map((e) => [e.id, e.size]))}
      <ZoneLayout
        groups={groupByZone(visible, categoryOf, { system: systemZone, keepEmpty: editingTab })}
        editing={editingTab}
      >
        {#snippet item(widget: Widget)}
          {@render frame(
            widget,
            { onpointerdown: () => {} },
            phone ? phoneSizes.get(widget.id) : widget.size,
            !phone,
          )}
        {/snippet}
      </ZoneLayout>
    {:else if phone}
      {@const tab = active}
      {@const layout = tab.mobileCustom ? resolveMobile(tab.mobile, visible) : deriveMobile(visible)}
      {@const phoneSizes = new Map(layout.map((e) => [e.id, e.size]))}
      <MobileList
        {layout}
        widgets={new Map(visible.map((w) => [w.id, w]))}
        editing={editingTab}
        custom={tab.mobileCustom}
        onchange={(layout) => onmobile(tab, layout)}
        onreset={() => resetMobile(tab)}
        label={(w) => titles.get(w.kind) ?? w.kind}
      >
        {#snippet item(widget: Widget)}
          <!-- On phones the phone layout's size decides the view. -->
          {@render frame(widget, { onpointerdown: () => {} }, phoneSizes.get(widget.id), false)}
        {/snippet}
      </MobileList>
    {:else}
      <Grid items={visible} editing={editingTab} {onlayout} label={(w) => titles.get(w.kind) ?? w.kind}>
        {#snippet item(widget: Widget, handle: { onpointerdown: (e: PointerEvent) => void })}
          {@render frame(widget, handle, widget.size, true)}
        {/snippet}
      </Grid>
    {/if}
    {#if visible.length === 0 && !folded && loaded}
      <p class="empty muted">
        {t(editingTab ? 'dashboard.empty_tab_editing' : 'dashboard.empty_tab_view')}
        {#if !editing && onedit && active.mine}
          <button
            class="btn"
            type="button"
            onclick={() => {
              onedit();
              dialog = { kind: 'add-widget' };
            }}>{t('dashboard.add_widget')}</button
          >
        {/if}
      </p>
    {/if}
    {#if editingTab}
      <div class="edit-actions" role="toolbar" aria-label={t('dashboard.edit_actions')}>
        <button class="btn btn-primary add" type="button" onclick={() => (dialog = { kind: 'add-widget' })}
          >{t('dashboard.add_widget')}</button
        >
        {#if active}
          {@const tab = active}
          <button class="btn" type="button" onclick={() => openBackground(tab)}
            >{t('dashboard.background_button')}</button
          >
          <button
            class="btn"
            type="button"
            aria-pressed={tab.zoned}
            title={t('dashboard.zones_hint')}
            onclick={() => toggleZones(tab)}>{t('dashboard.zones_button')}</button
          >
        {/if}
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
        {#if ondone}
          <button class="btn btn-primary done" type="button" onclick={ondone} aria-keyshortcuts="Escape"
            >✓ {t('dashboard.done')}</button
          >
        {/if}
      </div>
    {/if}
  {/if}
</section>

{#if detail}
  {@const widget = detail.widget}
  {@const notify = notifyAction(widget)}
  {@const mayEdit = !!onedit && (active?.mine ?? false) && hasSettings(widget.kind)}
  <DetailPanel
    title={titles.get(widget.kind) ?? widget.kind}
    opener={detail.opener}
    onclose={() => (detail = null)}
    actions={mayEdit || notify ? panelActions : undefined}
  >
    {@render content(widget)}
  </DetailPanel>
  {#snippet panelActions()}
    {#if notify}
      <button
        class="btn"
        type="button"
        onclick={() => {
          detail = null;
          notify();
        }}>{t('dashboard.notify_on_failure')}</button
      >
    {/if}
    {#if mayEdit}
      <button class="btn" type="button" onclick={() => editFromPanel(widget)}
        >{t('dashboard.edit_widget')}</button
      >
    {/if}
  {/snippet}
{/if}

{#if toast}
  <div class="toast" class:above-bar={editing} role="status">
    {toast.text}
    <button class="btn" type="button" onclick={undo}>{t('dashboard.undo')}</button>
  </div>
{/if}

{#if dialog?.kind === 'add-widget'}
  <WidgetPicker
    initialQuery={dialog.query}
    types={widgetTypes}
    icon={widgetIcon}
    onpick={addWidget}
    {compactExample}
    onclose={() => (dialog = null)}
    footer={pickerFooter}
  />
{:else if dialog?.kind === 'new-tab'}
  <NewTabDialog
    {uc}
    onclose={() => (dialog = null)}
    oncreated={(tab) => {
      tabs = [...tabs, tab];
      select(tab.id);
      offerTabId = tab.id;
      announce(t('dashboard.offer_question', { name: tab.name }));
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
{:else if dialog?.kind === 'rules'}
  {@const tab = dialog.tab}
  <TabRulesDialog
    {tab}
    onclose={() => (dialog = null)}
    onsubmit={async (rules) => {
      const updated = await uc.setRules(tab.id, rules);
      replaceTab({ ...tab, rules: updated.rules });
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
{:else if dialog?.kind === 'delete-tab'}
  {@const tab = dialog.tab}
  <Modal title={t('dashboard.delete_tab')} onclose={() => (dialog = null)}>
    <p>{t('dashboard.confirm_delete_tab', { name: tab.name })}</p>
    <div class="confirm-actions">
      <button class="btn" type="button" onclick={() => (dialog = null)}>{t('dashboard.cancel')}</button>
      <button class="btn btn-danger" type="button" onclick={() => confirmDeleteTab(tab)}
        >{t('dashboard.delete_tab')}</button
      >
    </div>
  </Modal>
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
  .confirm-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }
  .empty {
    text-align: center;
    padding: var(--space-6);
  }
  .add {
    align-self: center;
  }
  /* Stays at the bottom of the viewport while editing (#271). */
  .edit-actions {
    position: sticky;
    bottom: 0;
    z-index: 15;
    display: flex;
    gap: var(--space-2);
    justify-content: center;
    align-items: center;
    flex-wrap: wrap;
    margin-top: var(--space-3);
    padding: var(--space-2) var(--space-3) calc(var(--space-2) + env(safe-area-inset-bottom, 0px));
    background: color-mix(in srgb, var(--color-bg) 92%, transparent);
    backdrop-filter: blur(6px);
    border-top: 1px solid var(--color-border);
  }
  .edit-actions .done {
    margin-left: auto;
  }
  .edit-hint {
    margin: 0 0 var(--space-2);
    padding: var(--space-1) var(--space-3);
    border-left: 3px solid var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 10%, transparent);
    border-radius: var(--radius-small);
    font-size: 0.9rem;
  }
  .toast.above-bar {
    bottom: calc(var(--space-4) + 64px);
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
  :global(.flash) {
    outline: 3px solid var(--color-accent-text);
    outline-offset: 2px;
    border-radius: var(--radius);
    transition: outline-color 1.5s;
  }
</style>
