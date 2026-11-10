import type { Tab, Widget, WidgetConfig } from '../domain/dashboard';
import type { LayoutItem, Rect } from '../domain/layout';

/** Output port: persistence of tabs and widgets. */
export interface DashboardGateway {
  load(): Promise<Tab[]>;
  createTab(name: string): Promise<Tab>;
  renameTab(id: string, name: string): Promise<Tab>;
  deleteTab(id: string): Promise<void>;
  reorderTabs(ids: string[]): Promise<void>;
  addWidget(tabId: string, kind: string, rect: Rect, config: WidgetConfig): Promise<Widget>;
  configureWidget(id: string, config: WidgetConfig): Promise<Widget>;
  saveLayout(tabId: string, items: LayoutItem[]): Promise<void>;
  deleteWidget(id: string): Promise<void>;
}
