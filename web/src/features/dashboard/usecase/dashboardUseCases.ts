import {
  type Dashboard,
  reorder,
  type Tab,
  type TabRule,
  type TabTemplate,
  validateBackground,
  validateRules,
  validateTabName,
  type Widget,
  type WidgetConfig,
  type WidgetType,
} from '../domain/dashboard';
import { collides, findFreeSpot, type LayoutItem, sameLayout } from '../domain/layout';
import type { MobileItem } from '../domain/mobile';
import type { DashboardGateway } from './ports';

/** Error for input rejected before it reaches the backend. */
export class DashboardValidationError extends Error {
  override name = 'DashboardValidationError';
}

/** Application logic of the dashboard feature. */
export class DashboardUseCases {
  constructor(private readonly gateway: DashboardGateway) {}

  /** Loads all tabs with their widgets, sorted by position. */
  async load(): Promise<Tab[]> {
    return (await this.open()).tabs;
  }

  /** Loads the dashboard with the tab whose rules match now (#71). */
  async open(): Promise<Dashboard> {
    const { tabs, defaultTabId } = await this.gateway.load();
    return { tabs: [...tabs].sort((a, b) => a.position - b.position), defaultTabId };
  }

  /** Creates a tab after validating its name. */
  async createTab(name: string): Promise<Tab> {
    this.assertTabName(name);
    return this.gateway.createTab(name.trim());
  }

  /** Renames a tab after validating its name. */
  async renameTab(id: string, name: string): Promise<Tab> {
    this.assertTabName(name);
    return this.gateway.renameTab(id, name.trim());
  }

  /** Deletes a tab including its widgets. */
  deleteTab(id: string): Promise<void> {
    return this.gateway.deleteTab(id);
  }

  /** Moves a tab within the tab bar and persists the new order. */
  async moveTab(tabs: readonly Tab[], from: number, to: number): Promise<Tab[]> {
    const ordered = reorder(tabs, from, to).map((t, position) => ({ ...t, position }));
    await this.gateway.reorderTabs(ordered.map((t) => t.id));
    return ordered;
  }

  /** Adds a widget of the given type at the first free spot of the tab. */
  addWidget(
    tab: Tab,
    type: WidgetType,
    config: WidgetConfig = structuredClone(type.defaultConfig ?? {}),
  ): Promise<Widget> {
    const rect = findFreeSpot(tab.widgets, type.defaultSize.w, type.defaultSize.h);
    return this.gateway.addWidget(tab.id, type.kind, rect, config);
  }

  /** Stores a new configuration for a widget. */
  configureWidget(id: string, config: WidgetConfig): Promise<Widget> {
    return this.gateway.configureWidget(id, config);
  }

  /** Persists a layout if it differs from the previous one. */
  async saveLayout(tabId: string, before: readonly LayoutItem[], after: LayoutItem[]): Promise<void> {
    if (sameLayout(before, after)) return;
    await this.gateway.saveLayout(tabId, after);
  }

  /** Copies a widget (kind, size and configuration) to the first free spot of its tab. */
  duplicateWidget(tab: Tab, widget: Widget): Promise<Widget> {
    const rect = findFreeSpot(tab.widgets, widget.w, widget.h);
    return this.gateway.addWidget(tab.id, widget.kind, rect, structuredClone(widget.config));
  }

  /**
   * Re-creates a deleted widget with its kind, size and configuration
   * (undo, #93): at its old position, or the first free spot if that is
   * taken now. The new widget has a new ID.
   */
  restoreWidget(tab: Tab, widget: Widget): Promise<Widget> {
    const others = tab.widgets.filter((w) => w.id !== widget.id);
    const free = !others.some((w) => collides(w, widget));
    const rect = free
      ? { x: widget.x, y: widget.y, w: widget.w, h: widget.h }
      : findFreeSpot(others, widget.w, widget.h);
    return this.gateway.addWidget(tab.id, widget.kind, rect, structuredClone(widget.config));
  }

  /** Sets the background image of a tab ("" = global background). */
  async setBackground(tabId: string, background: string): Promise<Tab> {
    const problem = validateBackground(background);
    if (problem) throw new DashboardValidationError(problem);
    return this.gateway.setBackground(tabId, background.trim());
  }

  /** Sets the rules that open a tab by default (#71). */
  async setRules(tabId: string, rules: TabRule[]): Promise<Tab> {
    const clean = rules.map((r) => ({
      days: [...new Set(r.days)].sort((a, b) => a - b),
      from: r.from.trim(),
      to: r.to.trim(),
      networks: r.networks.map((n) => n.trim()).filter((n) => n !== ''),
    }));
    const problem = validateRules(clean);
    if (problem) throw new DashboardValidationError(problem);
    return this.gateway.setRules(tabId, clean);
  }

  /** Stores the phone layout of a tab (#92). */
  saveMobileLayout(tabId: string, items: MobileItem[]): Promise<Tab> {
    return this.gateway.saveMobileLayout(tabId, items);
  }

  /** Derives the phone layout from the desktop grid again (#92). */
  resetMobileLayout(tabId: string): Promise<Tab> {
    return this.gateway.resetMobileLayout(tabId);
  }

  listTemplates(): Promise<TabTemplate[]> {
    return this.gateway.listTemplates();
  }

  /** Saves the widgets of a tab as a template. */
  async saveTemplate(tabId: string, name: string): Promise<TabTemplate> {
    this.assertTabName(name);
    return this.gateway.saveTemplate(tabId, name.trim());
  }

  deleteTemplate(id: string): Promise<void> {
    return this.gateway.deleteTemplate(id);
  }

  /** Creates a new tab from a template; the template name is used if name is empty. */
  createTabFromTemplate(templateId: string, name: string): Promise<Tab> {
    if (name.trim() !== '') this.assertTabName(name);
    return this.gateway.createTabFromTemplate(templateId, name.trim());
  }

  /** Removes a widget. */
  deleteWidget(id: string): Promise<void> {
    return this.gateway.deleteWidget(id);
  }

  private assertTabName(name: string): void {
    const problem = validateTabName(name);
    if (problem) throw new DashboardValidationError(problem);
  }
}
