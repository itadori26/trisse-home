import {
  reorder,
  type Tab,
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

  /** Removes a widget. */
  deleteWidget(id: string): Promise<void> {
    return this.gateway.deleteWidget(id);
  }

  private assertTabName(name: string): void {
    const problem = validateTabName(name);
    if (problem) throw new DashboardValidationError(problem);
  }
}
