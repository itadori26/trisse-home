import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  DashboardDTO,
  TabDTO,
  TemplateDTO,
  BuiltinTemplateDTO,
  WidgetDTO,
} from '../../../../shared/api-client/generated/dashboard';
import type {
  AddWidgetDTO,
  LayoutDTO,
  SaveTemplateDTO,
  BuiltinTabDTO,
  TabBackgroundDTO,
  TabNameDTO,
  TabOrderDTO,
  TabRulesDTO,
  TabZonesDTO,
  MobileLayoutDTO,
  WidgetConfigDTO,
} from '../../../../shared/api-client/generated/dashboard-requests';
import type {
  BuiltinTemplate,
  Dashboard,
  Tab,
  TabRule,
  TabTemplate,
  TemplateLocation,
  Widget,
  WidgetConfig,
} from '../../domain/dashboard';
import type { LayoutItem, Rect } from '../../domain/layout';
import { type MobileItem, type MobileSize, sizeCells } from '../../domain/mobile';
import { parseSize, type WidgetSize } from '../../domain/sizes';
import { parseZone, type Zone } from '../../domain/zones';
import type { DashboardGateway } from '../../usecase/ports';

const toWidget = (dto: WidgetDTO): Widget => {
  const size = parseSize(dto.size);
  const zone = parseZone(dto.zone);
  return {
    id: dto.id,
    tabId: dto.tabId,
    kind: dto.kind,
    config: dto.config ?? {},
    x: dto.x,
    y: dto.y,
    w: dto.w,
    h: dto.h,
    ...(size ? { size } : {}),
    ...(zone ? { zone } : {}),
  };
};

const toTab = (dto: TabDTO): Tab => ({
  id: dto.id,
  name: dto.name,
  position: dto.position,
  background: dto.background ?? '',
  rules: (dto.rules ?? []).map((r) => ({
    days: r.days ?? [],
    from: r.from,
    to: r.to,
    networks: r.networks ?? [],
  })),
  widgets: dto.widgets.map(toWidget),
  mobile: (dto.mobile ?? []).map((m) => ({
    id: m.id,
    hidden: m.hidden,
    rows: m.rows,
    ...(sizeCells(m.size as MobileSize) ? { size: m.size as MobileSize } : {}),
  })),
  // Older servers do not send it; every tab was the viewer's own then.
  mine: dto.mine ?? true,
  mobileCustom: dto.mobileCustom ?? false,
  zoned: dto.zoned ?? false,
  ...(dto.sample ? { sample: true } : {}),
});

/** DashboardGateway backed by the REST API. */
export class DashboardApi implements DashboardGateway {
  constructor(private readonly http: HttpClient) {}

  async load(): Promise<Dashboard> {
    const dto = await this.http.get<DashboardDTO>('/api/dashboard');
    return {
      tabs: dto.tabs.map(toTab),
      defaultTabId: dto.defaultTabId ?? '',
      clientNetwork: dto.clientNetwork,
    };
  }

  async createTab(name: string): Promise<Tab> {
    const body: TabNameDTO = { name };
    return toTab(await this.http.post<TabDTO>('/api/dashboard/tabs', body));
  }

  async renameTab(id: string, name: string): Promise<Tab> {
    const body: TabNameDTO = { name };
    return toTab(await this.http.patch<TabDTO>(`/api/dashboard/tabs/${encodeURIComponent(id)}`, body));
  }

  deleteTab(id: string): Promise<void> {
    return this.http.delete(`/api/dashboard/tabs/${encodeURIComponent(id)}`);
  }

  reorderTabs(ids: string[]): Promise<void> {
    const body: TabOrderDTO = { tabIds: ids };
    return this.http.put('/api/dashboard/tabs/order', body);
  }

  async addWidget(
    tabId: string,
    kind: string,
    rect: Rect,
    config: WidgetConfig,
    size?: WidgetSize,
  ): Promise<Widget> {
    const body: AddWidgetDTO = { kind, ...rect, config, ...(size ? { size } : {}) };
    return toWidget(
      await this.http.post<WidgetDTO>(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/widgets`, body),
    );
  }

  async configureWidget(id: string, config: WidgetConfig): Promise<Widget> {
    const body: WidgetConfigDTO = { config };
    return toWidget(
      await this.http.patch<WidgetDTO>(`/api/dashboard/widgets/${encodeURIComponent(id)}`, body),
    );
  }

  saveLayout(tabId: string, items: LayoutItem[]): Promise<void> {
    // The size is always sent, so a tile made free again loses it on the server too.
    const body: LayoutDTO = {
      items: items.map(({ id, x, y, w, h, size }) => ({ id, x, y, w, h, size: size ?? '' })),
    };
    return this.http.put(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/layout`, body);
  }

  deleteWidget(id: string): Promise<void> {
    return this.http.delete(`/api/dashboard/widgets/${encodeURIComponent(id)}`);
  }

  async setBackground(tabId: string, background: string): Promise<Tab> {
    const body: TabBackgroundDTO = { background };
    return toTab(
      await this.http.put<TabDTO>(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/background`, body),
    );
  }

  async setZones(tabId: string, enabled: boolean): Promise<Tab> {
    const body: TabZonesDTO = { enabled };
    return toTab(await this.http.put<TabDTO>(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/zones`, body));
  }

  moveToZone(widget: Widget, zone: Zone | undefined): Promise<void> {
    // Only this widget is listed, so the others keep their places and zones.
    const { id, x, y, w, h } = widget;
    const body: LayoutDTO = { items: [{ id, x, y, w, h, zone: zone ?? '' }] };
    return this.http.put(`/api/dashboard/tabs/${encodeURIComponent(widget.tabId)}/layout`, body);
  }

  async setRules(tabId: string, rules: TabRule[]): Promise<Tab> {
    const body: TabRulesDTO = { rules };
    return toTab(await this.http.put<TabDTO>(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/rules`, body));
  }

  async saveMobileLayout(tabId: string, items: MobileItem[]): Promise<Tab> {
    const body: MobileLayoutDTO = {
      items: items.map(({ id, hidden, rows, size }) => ({ id, hidden, rows, ...(size ? { size } : {}) })),
    };
    return toTab(
      await this.http.put<TabDTO>(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/mobile-layout`, body),
    );
  }

  async resetMobileLayout(tabId: string): Promise<Tab> {
    return toTab(
      await this.http.delete<TabDTO>(`/api/dashboard/tabs/${encodeURIComponent(tabId)}/mobile-layout`),
    );
  }

  async listTemplates(): Promise<TabTemplate[]> {
    return (await this.http.get<TemplateDTO[]>('/api/dashboard/templates')).map((t) => ({ ...t }));
  }

  async saveTemplate(tabId: string, name: string): Promise<TabTemplate> {
    const body: SaveTemplateDTO = { tabId, name };
    return { ...(await this.http.post<TemplateDTO>('/api/dashboard/templates', body)) };
  }

  deleteTemplate(id: string): Promise<void> {
    return this.http.delete(`/api/dashboard/templates/${encodeURIComponent(id)}`);
  }

  async listBuiltinTemplates(): Promise<BuiltinTemplate[]> {
    const list = await this.http.get<BuiltinTemplateDTO[]>('/api/dashboard/templates/builtin');
    return list.map((t) => ({
      id: t.id,
      title: t.title,
      needs: t.needs.filter((n): n is 'location' => n === 'location'),
      widgets: t.widgets.map((w) => ({ ...w })),
    }));
  }

  async createTabFromBuiltin(id: string, name: string, location?: TemplateLocation): Promise<Tab> {
    const body: BuiltinTabDTO = { name, location: location ? { placeId: '', ...location } : undefined };
    return toTab(
      await this.http.post<TabDTO>(`/api/dashboard/templates/builtin/${encodeURIComponent(id)}/tabs`, body),
    );
  }

  discardSample(): Promise<void> {
    return this.http.delete('/api/dashboard/sample');
  }

  async keepSample(): Promise<void> {
    await this.http.post('/api/dashboard/sample/keep', {});
  }

  async createTabFromTemplate(templateId: string, name: string): Promise<Tab> {
    const body: TabNameDTO = { name };
    return toTab(
      await this.http.post<TabDTO>(`/api/dashboard/templates/${encodeURIComponent(templateId)}/tabs`, body),
    );
  }
}
