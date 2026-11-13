import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  DiscoveryDTO,
  IncidentDTO,
  LatencyPointDTO,
  ServiceDTO,
  UptimeDTO,
} from '../../../../shared/api-client/generated/services';
import type {
  DiscoveryInputDTO,
  OrderDTO,
  ServiceInputDTO,
} from '../../../../shared/api-client/generated/services-requests';
import type { Discovery, DiscoverySettings } from '../../domain/discovery';
import type { LatencyPoint, Service, ServiceInput } from '../../domain/service';
import type { Incident, ServiceUptime } from '../../domain/uptime';
import type { ServicesGateway } from '../../usecase/servicesUseCases';

const ZERO_TIME = '0001-01-01T00:00:00Z';

const toService = (d: ServiceDTO): Service => ({
  id: d.id,
  name: d.name,
  url: d.url,
  checkUrl: d.checkUrl,
  icon: d.icon,
  description: d.description,
  skipTlsVerify: d.skipTlsVerify,
  mac: d.mac,
  position: d.position,
  state: d.state,
  httpStatus: d.httpStatus,
  latencyMs: d.latencyMs,
  checkedAt: d.checkedAt && d.checkedAt !== ZERO_TIME ? new Date(d.checkedAt) : null,
  error: d.error,
  certExpiresAt: d.certExpiresAt ? new Date(d.certExpiresAt) : null,
  certExpiresSoon: d.certExpiresSoon,
  container: d.container,
  hidden: d.hidden,
});

const toDiscovery = (d: DiscoveryDTO): Discovery => ({
  available: d.available,
  enabled: d.enabled,
  intervalMinutes: d.intervalMinutes,
  rule: d.rule,
  lastSync: d.lastSync ? { ...d.lastSync, at: new Date(d.lastSync.at) } : null,
});

const toDTO = (i: ServiceInput): ServiceInputDTO => ({ ...i, mac: i.mac.trim() });

/** ServicesGateway backed by the REST API. */
export class ServicesApi implements ServicesGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Service[]> {
    return (await this.http.get<ServiceDTO[]>('/api/services')).map(toService);
  }
  async create(input: ServiceInput): Promise<Service> {
    return toService(await this.http.post<ServiceDTO>('/api/services', toDTO(input)));
  }
  async update(id: string, input: ServiceInput): Promise<Service> {
    return toService(
      await this.http.put<ServiceDTO>(`/api/services/${encodeURIComponent(id)}`, toDTO(input)),
    );
  }
  delete(id: string): Promise<void> {
    return this.http.delete(`/api/services/${encodeURIComponent(id)}`);
  }
  reorder(ids: string[]): Promise<void> {
    const body: OrderDTO = { ids };
    return this.http.put('/api/services/order', body);
  }
  async checkNow(): Promise<Service[]> {
    return (await this.http.post<ServiceDTO[]>('/api/services/check')).map(toService);
  }

  async history(hours: number): Promise<Record<string, LatencyPoint[]>> {
    const dto = await this.http.get<Record<string, LatencyPointDTO[]>>(
      `/api/services/history?hours=${hours}`,
    );
    return Object.fromEntries(
      Object.entries(dto).map(([id, ps]) => [
        id,
        ps.map((p) => ({ at: new Date(p.at), latencyMs: p.latencyMs, online: p.online })),
      ]),
    );
  }

  wake(id: string): Promise<void> {
    return this.http.post(`/api/services/${encodeURIComponent(id)}/wake`);
  }

  async uptime(): Promise<ServiceUptime[]> {
    return (await this.http.get<UptimeDTO[]>('/api/services/uptime')).map((d) => ({
      serviceId: d.serviceId,
      day: d.day ?? null,
      week: d.week ?? null,
      month: d.month ?? null,
      quarter: d.quarter ?? null,
      days: (d.days ?? []).map((x) => x ?? null),
    }));
  }

  async incidents(serviceId: string, limit: number): Promise<Incident[]> {
    const q = new URLSearchParams({ service: serviceId, limit: String(limit) });
    return (await this.http.get<IncidentDTO[]>(`/api/services/incidents?${q}`)).map((d) => ({
      ...d,
      start: new Date(d.start),
      end: d.end ? new Date(d.end) : null,
    }));
  }

  async discovery(): Promise<Discovery> {
    return toDiscovery(await this.http.get<DiscoveryDTO>('/api/services/discovery'));
  }

  async saveDiscovery(settings: DiscoverySettings): Promise<Discovery> {
    const body: DiscoveryInputDTO = settings;
    return toDiscovery(await this.http.put<DiscoveryDTO>('/api/services/discovery', body));
  }

  async syncDiscovery(): Promise<Discovery> {
    return toDiscovery(await this.http.post<DiscoveryDTO>('/api/services/discovery/sync'));
  }
}
