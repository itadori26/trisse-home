import type { HttpClient } from '../../../../shared/api-client/http';
import type { LatencyPointDTO, ServiceDTO } from '../../../../shared/api-client/generated/services';
import type { OrderDTO, ServiceInputDTO } from '../../../../shared/api-client/generated/services-requests';
import type { LatencyPoint, Service, ServiceInput } from '../../domain/service';
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
  position: d.position,
  state: d.state,
  httpStatus: d.httpStatus,
  latencyMs: d.latencyMs,
  checkedAt: d.checkedAt && d.checkedAt !== ZERO_TIME ? new Date(d.checkedAt) : null,
  error: d.error,
  certExpiresAt: d.certExpiresAt ? new Date(d.certExpiresAt) : null,
  certExpiresSoon: d.certExpiresSoon,
});

const toDTO = (i: ServiceInput): ServiceInputDTO => ({ ...i });

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
}
