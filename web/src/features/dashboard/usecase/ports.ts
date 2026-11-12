import type { Dashboard, Tab, TabRule, TabTemplate, Widget, WidgetConfig } from '../domain/dashboard';
import type { LayoutItem, Rect } from '../domain/layout';

/** Output port: persistence of tabs and widgets. */
export interface DashboardGateway {
  load(): Promise<Dashboard>;
  createTab(name: string): Promise<Tab>;
  renameTab(id: string, name: string): Promise<Tab>;
  deleteTab(id: string): Promise<void>;
  reorderTabs(ids: string[]): Promise<void>;
  addWidget(tabId: string, kind: string, rect: Rect, config: WidgetConfig): Promise<Widget>;
  configureWidget(id: string, config: WidgetConfig): Promise<Widget>;
  saveLayout(tabId: string, items: LayoutItem[]): Promise<void>;
  deleteWidget(id: string): Promise<void>;
  setBackground(tabId: string, background: string): Promise<Tab>;
  setRules(tabId: string, rules: TabRule[]): Promise<Tab>;
  listTemplates(): Promise<TabTemplate[]>;
  saveTemplate(tabId: string, name: string): Promise<TabTemplate>;
  deleteTemplate(id: string): Promise<void>;
  createTabFromTemplate(templateId: string, name: string): Promise<Tab>;
}
