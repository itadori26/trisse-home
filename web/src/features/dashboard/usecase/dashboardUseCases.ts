import {
  reorder,
  type Tab,
  type TabTemplate,
  validateBackground,
  validateTabName,
  type Widget,
  type WidgetConfig,
  type WidgetType,
} from '../domain/dashboard';
import { findFreeSpot, type LayoutItem, sameLayout } from '../domain/layout';
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
    const tabs = await this.gateway.load();
    return [...tabs].sort((a, b) => a.position - b.position);
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
  addWidget(tab: Tab, type: WidgetType, config: WidgetConfig = {}): Promise<Widget> {
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

  /** Sets the background image of a tab ("" = global background). */
  async setBackground(tabId: string, background: string): Promise<Tab> {
    const problem = validateBackground(background);
    if (problem) throw new DashboardValidationError(problem);
    return this.gateway.setBackground(tabId, background.trim());
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
