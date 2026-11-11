import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  DashboardDTO,
  TabDTO,
  TemplateDTO,
  WidgetDTO,
} from '../../../../shared/api-client/generated/dashboard';
import type {
  AddWidgetDTO,
  LayoutDTO,
  SaveTemplateDTO,
  TabBackgroundDTO,
  TabNameDTO,
  TabOrderDTO,
  WidgetConfigDTO,
} from '../../../../shared/api-client/generated/dashboard-requests';
import type { Tab, TabTemplate, Widget, WidgetConfig } from '../../domain/dashboard';
import type { LayoutItem, Rect } from '../../domain/layout';
import type { DashboardGateway } from '../../usecase/ports';

const toWidget = (dto: WidgetDTO): Widget => ({
  id: dto.id,
  tabId: dto.tabId,
  kind: dto.kind,
  config: dto.config ?? {},
  x: dto.x,
  y: dto.y,
  w: dto.w,
  h: dto.h,
});

const toTab = (dto: TabDTO): Tab => ({
  id: dto.id,
  name: dto.name,
  position: dto.position,
  background: dto.background ?? '',
  widgets: dto.widgets.map(toWidget),
});

/** DashboardGateway backed by the REST API. */
export class DashboardApi implements DashboardGateway {
  constructor(private readonly http: HttpClient) {}

  async load(): Promise<Tab[]> {
    const dto = await this.http.get<DashboardDTO>('/api/dashboard');
    return dto.tabs.map(toTab);
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

  async addWidget(tabId: string, kind: string, rect: Rect, config: WidgetConfig): Promise<Widget> {
    const body: AddWidgetDTO = { kind, ...rect, config };
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
    const body: LayoutDTO = { items: items.map(({ id, x, y, w, h }) => ({ id, x, y, w, h })) };
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

  async createTabFromTemplate(templateId: string, name: string): Promise<Tab> {
    const body: TabNameDTO = { name };
    return toTab(
      await this.http.post<TabDTO>(`/api/dashboard/templates/${encodeURIComponent(templateId)}/tabs`, body),
    );
  }
}
