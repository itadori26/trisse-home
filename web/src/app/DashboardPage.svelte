<script lang="ts">
  import type { Appearance } from '../features/appearance';
  import type { BackgroundScope } from '../features/appearance/domain/backgrounds';
  import { DashboardView, type Tab, type Widget, type WidgetConfig } from '../features/dashboard';
  import type { Container } from './container';
  import LazyComponent from './LazyComponent.svelte';
  import { definitionFor, iconFor, widgetDefinitions } from './widgets';
  import { lazy } from './widgetKit';
  import { useOpenSettings } from '../shared/ui/settingsLink';
  import { type StatusSource, uniqueSources } from '../features/dashboard/domain/attentionLine';
  import { categoryShown, type StartProfile } from '../features/appearance/domain/appearance';
  import { COMPACT_REFRESH_MS } from '../features/compact/domain/compact';
  import { usePolling } from '../shared/refresh';
  import { SvelteSet } from 'svelte/reactivity';

  /** Props of the dashboard page. */
  interface Props {
    container: Container;
    editing: boolean;
    /** What the start page shows (#281). */
    profile?: StartProfile;
    /** Widget kinds enabled in the backend. */
    widgetKinds: Set<string>;
    /** Element id in the top bar that hosts the tab bar (#167). */
    tabBarTarget?: string;
    /** Reports tabs and the active tab to the shell. */
    onstate?: (tabs: Tab[], activeId: string) => void;
    /** Starts edit mode; omitted for guests (#276). */
    onedit?: () => void;
    /** Finishes edit mode (#271). */
    ondone?: () => void;
    /** The global background, for the background picker (#277). */
    globalBackground?: string;
    /** Reports a global background saved from the dashboard. */
    onappearance?: (a: Appearance) => void;
  }
  let {
    container,
    editing,
    profile = 'homelab',
    widgetKinds,
    tabBarTarget,
    onstate,
    onedit,
    ondone,
    globalBackground = '',
    onappearance,
  }: Props = $props();

  let view = $state<{
    selectTabAt: (index: number) => void;
    undo: () => void;
    redo: () => void;
    showTabBackground: (tabId: string, background: string) => void;
    selectTab: (id: string) => void;
    openWidgetSettings: (tabId: string, widgetId: string) => void;
    openAddWidget: (query?: string) => void;
    addTab: (tab: Tab) => void;
    reload: (selectId?: string) => Promise<void>;
  }>();
  // The picker is loaded only when someone opens it (#116).
  let pickerTab = $state<Tab | null>(null);
  const loadPicker = () => import('../features/appearance/ui/BackgroundPicker.svelte');

  async function saveBackground(tab: Tab, value: string, scope: BackgroundScope, clearTab: boolean) {
    if (scope === 'tab') {
      const updated = await container.dashboard.setBackground(tab.id, value);
      view?.showTabBackground(tab.id, updated.background);
      return;
    }
    onappearance?.(await container.appearance.setBackground(value));
    if (clearTab) {
      await container.dashboard.setBackground(tab.id, '');
      view?.showTabBackground(tab.id, '');
    }
  }

  /** Selects the n-th tab (0-based). */
  export function selectTabAt(index: number) {
    view?.selectTabAt(index);
  }

  /** Selects a tab by id and opens the widget picker, from the search (#476). */
  export function selectTab(id: string) {
    view?.selectTab(id);
  }
  /** Opens a widget's settings on its tab, from the attention list (#592). */
  export function openWidgetSettings(tabId: string, widgetId: string) {
    view?.openWidgetSettings(tabId, widgetId);
  }
  /** Shows a tab created from a template on first start (#589). */
  export function addTab(tab: Tab) {
    view?.addTab(tab);
  }
  /** Loads the tabs again after the template choice replaced the example tabs (#670). */
  export function reload(selectId?: string) {
    return view?.reload(selectId);
  }
  export function openAddWidget(query = '') {
    view?.openAddWidget(query);
  }

  /** Undo/redo in edit mode (#93). */
  export function undo() {
    view?.undo();
  }
  export function redo() {
    view?.redo();
  }

  // One component shows the compact state of every module (#595).
  const compactView = lazy(
    () => import('../features/compact/ui/CompactView.svelte'),
    (_config, c) => ({ compact: c.compact, module: '', title: '' }),
  );

  /** The module of a widget's compact view; none in local mode, which has no compact endpoints. */
  function compactModule(widget: Widget): string | undefined {
    if (container.localMode) return undefined;
    return definitionFor(widget.kind)?.compact?.(widget.config);
  }

  // Zoned tabs hide widgets with nothing to show (#619). Whether a widget is
  // empty is part of its module's compact state, so the modules of the
  // active zoned tab are asked in the background.
  let shown = $state<Tab | undefined>();
  const emptyModules = new SvelteSet<string>();
  const zonedModules = $derived(
    shown?.zoned ? [...new Set(shown.widgets.flatMap((w) => compactModule(w) ?? []))].sort() : [],
  );
  const zonedKey = $derived(zonedModules.join(','));
  async function loadEmpty(modules: string[]) {
    const results = await Promise.allSettled(modules.map((m) => container.compact.state(m)));
    results.forEach((r, i) => {
      const module = modules[i];
      if (!module || r.status !== 'fulfilled') return;
      if (r.value.empty) emptyModules.add(module);
      else emptyModules.delete(module);
    });
  }
  const poll = usePolling();
  $effect(() => {
    if (zonedKey === '') return;
    const modules = zonedKey.split(',');
    return poll(() => loadEmpty(modules), COMPACT_REFRESH_MS);
  });
  function reportState(tabs: Tab[], activeId: string) {
    shown = tabs.find((t) => t.id === activeId);
    onstate?.(tabs, activeId);
  }

  // The lines above the grid load with the first tab that shows them (#593).
  const loadOverview = () => import('./DashboardOverview.svelte');

  /** The status widgets of a tab with the module of their compact state. */
  function statusSources(widgets: Widget[]): StatusSource[] {
    return uniqueSources(
      widgets.flatMap((w) => {
        const module = compactModule(w);
        return module ? [{ widgetId: w.id, module, title: definitionFor(w.kind)?.type.title ?? w.kind }] : [];
      }),
    );
  }

  // Notifications are set up by admins in the settings (#594).
  const openSettings = useOpenSettings();
  function notifyAction(widget: Widget): (() => void) | undefined {
    const open = openSettings;
    if (!open || !container.auth.admin || !definitionFor(widget.kind)?.alerts) return undefined;
    return () => open('notifications');
  }

  // The "everyday" profile hides home-server and network widgets from the
  // picker (#281); widgets already placed stay.
  const types = $derived(
    widgetDefinitions
      .filter((d) => widgetKinds.has(d.type.kind) && categoryShown(profile, d.type.category))
      .map((d) => d.type),
  );
</script>

<DashboardView
  bind:this={view}
  uc={container.dashboard}
  {editing}
  widgetTypes={types}
  widgetIcon={iconFor}
  hasSettings={(kind) => !!definitionFor(kind)?.settings}
  chromeless={(kind) => !!definitionFor(kind)?.chromeless}
  {tabBarTarget}
  onstate={reportState}
  {onedit}
  {ondone}
  onbackground={container.auth.guest ? undefined : (tab) => (pickerTab = tab)}
  pickerFooter={container.localMode ? fullVersionHint : undefined}
  hasCompact={(widget) => compactModule(widget) !== undefined}
  detailPanel={!container.auth.kiosk}
  {notifyAction}
  statusWidget={(widget) => !!definitionFor(widget.kind)?.alerts && compactModule(widget) !== undefined}
  compactExample={container.localMode ? undefined : (kind) => definitionFor(kind)?.compactExample?.()}
  systemZone={container.localMode || container.auth.admin}
  emptyWidget={(widget) => {
    const module = compactModule(widget);
    return module !== undefined && emptyModules.has(module);
  }}
>
  {#snippet content(widget: Widget)}
    {@const def = definitionFor(widget.kind)}
    {#if def}
      <LazyComponent
        renderer={def.view}
        props={def.view.props(widget.config, container)}
        scheduler={container.refresh}
        label={def.type.title}
      />
    {/if}
  {/snippet}
  {#snippet overview(o: {
    tab: Tab;
    status: Widget[];
    collapsed: boolean;
    toggle: () => void;
    view: (widgetId: string) => void;
  })}
    {@const sources = statusSources(o.status)}
    {#if !container.localMode}
      {#await loadOverview() then { default: DashboardOverview }}
        <DashboardOverview
          compact={container.compact}
          tabId={o.tab.id}
          {sources}
          collapsed={o.collapsed}
          ontoggle={o.toggle}
          onview={o.view}
        />
      {/await}
    {/if}
  {/snippet}
  {#snippet compact(widget: Widget, opts: { strip: boolean })}
    {@const def = definitionFor(widget.kind)}
    {@const module = compactModule(widget)}
    {#if def && module}
      <LazyComponent
        renderer={compactView}
        props={{
          ...compactView.props(widget.config, container),
          module,
          title: def.type.title,
          strip: opts.strip,
          // Guests and kiosk displays only look (#50).
          actions: !container.auth.guest,
        }}
        scheduler={container.refresh}
        label={def.type.title}
      />
    {/if}
  {/snippet}
  {#snippet settings(widget: Widget, save: (config: WidgetConfig) => Promise<void>)}
    {@const def = definitionFor(widget.kind)}
    {#if def?.settings}
      <LazyComponent
        renderer={def.settings}
        props={{ ...def.settings.props(widget.config, container), onsave: save }}
        scheduler={container.refresh}
        showStatus={false}
      />
    {/if}
  {/snippet}
</DashboardView>

{#snippet fullVersionHint()}
  <!-- Local mode of the extension (#291): say what a server adds. -->
  {#await import('./local/FullVersionHint.svelte') then { default: FullVersionHint }}
    <FullVersionHint variant="short" />
  {/await}
{/snippet}

{#if pickerTab}
  {@const tab = pickerTab}
  {#await loadPicker() then { default: BackgroundPicker }}
    <BackgroundPicker
      appearance={container.appearance}
      globalValue={globalBackground}
      tab={{ name: tab.name, background: tab.background }}
      onsave={(value, scope, clearTab) => saveBackground(tab, value, scope, clearTab)}
      onclose={() => (pickerTab = null)}
    />
  {/await}
{/if}
