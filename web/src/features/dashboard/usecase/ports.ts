import type {
  Dashboard,
  Tab,
  TabRule,
  TabTemplate,
  Widget,
  WidgetConfig,
  BuiltinTemplate,
  TemplateLocation,
} from '../domain/dashboard';
import type { LayoutItem, Rect } from '../domain/layout';
import type { MobileItem } from '../domain/mobile';
import type { WidgetSize } from '../domain/sizes';
import type { Zone } from '../domain/zones';

/** Output port: persistence of tabs and widgets. */
export interface DashboardGateway {
  load(): Promise<Dashboard>;
  createTab(name: string): Promise<Tab>;
  renameTab(id: string, name: string): Promise<Tab>;
  deleteTab(id: string): Promise<void>;
  reorderTabs(ids: string[]): Promise<void>;
  /** Adds a widget; a size (#595) makes it a sized tile instead of a free one. */
  addWidget(
    tabId: string,
    kind: string,
    rect: Rect,
    config: WidgetConfig,
    size?: WidgetSize,
  ): Promise<Widget>;
  configureWidget(id: string, config: WidgetConfig): Promise<Widget>;
  saveLayout(tabId: string, items: LayoutItem[]): Promise<void>;
  deleteWidget(id: string): Promise<void>;
  setBackground(tabId: string, background: string): Promise<Tab>;
  setRules(tabId: string, rules: TabRule[]): Promise<Tab>;
  /** Lays a tab out in zones or in the free grid again (#619). */
  setZones(tabId: string, enabled: boolean): Promise<Tab>;
  /** Moves a widget to a zone; undefined lets its kind decide again. */
  moveToZone(widget: Widget, zone: Zone | undefined): Promise<void>;
  saveMobileLayout(tabId: string, items: MobileItem[]): Promise<Tab>;
  resetMobileLayout(tabId: string): Promise<Tab>;
  listTemplates(): Promise<TabTemplate[]>;
  saveTemplate(tabId: string, name: string): Promise<TabTemplate>;
  deleteTemplate(id: string): Promise<void>;
  createTabFromTemplate(templateId: string, name: string): Promise<Tab>;
  /** Built-in templates (#589); absent where there is no server. */
  listBuiltinTemplates?(): Promise<BuiltinTemplate[]>;
  createTabFromBuiltin?(id: string, name: string, location?: TemplateLocation): Promise<Tab>;
  /** Removes the example tabs of the setup (#670); absent where there is no server. */
  discardSample?(): Promise<void>;
  /** Keeps the example tabs of the setup as ordinary tabs (#670). */
  keepSample?(): Promise<void>;
}
