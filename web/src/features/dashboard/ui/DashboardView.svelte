<script lang="ts">
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

  function onlayout(items: LayoutItem[]) {
    const tab = active;
    if (!tab) return;
    const before = tab.widgets;
    const byId = new Map(items.map((i) => [i.id, i]));
    replaceTab({ ...tab, widgets: tab.widgets.map((w) => ({ ...w, ...byId.get(w.id) })) });
    run(() => uc.saveLayout(tab.id, before, items));
  }

  function addWidget(type: WidgetType) {
    const tab = active;
    if (!tab) return;
    dialog = null;
    run(async () => {
      const widget = await uc.addWidget(tab, type);
      replaceTab({ ...tab, widgets: [...tab.widgets, widget] });
      if (hasSettings(type.kind)) dialog = { kind: 'settings', widget };
    });
  }

  function duplicateWidget(widget: Widget) {
    const tab = tabs.find((t) => t.id === widget.tabId);
    if (!tab) return;
    run(async () => {
      const copy = await uc.duplicateWidget(tab, widget);
      replaceTab({ ...tab, widgets: [...tab.widgets, copy] });
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

  function removeWidget(widget: Widget) {
    if (!confirm(t('dashboard.confirm_remove_widget', { name: titles.get(widget.kind) ?? widget.kind })))
      return;
    run(async () => {
      await uc.deleteWidget(widget.id);
      const tab = tabs.find((t) => t.id === widget.tabId);
      if (tab) replaceTab({ ...tab, widgets: tab.widgets.filter((w) => w.id !== widget.id) });
    });
  }

  async function saveConfig(widget: Widget, config: WidgetConfig) {
    const updated = await uc.configureWidget(widget.id, config);
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
      <button class="btn btn-primary add" type="button" onclick={() => (dialog = { kind: 'add-widget' })}
        >{t('dashboard.add_widget')}</button
      >
    {/if}
  {/if}
</section>

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
</style>
