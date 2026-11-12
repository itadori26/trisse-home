import type { HttpClient } from '../../../../shared/api-client/http';
import type { InstanceDTO, OverviewDTO } from '../../../../shared/api-client/generated/arr';
import type { InstanceInputDTO } from '../../../../shared/api-client/generated/arr-requests';
import type { ArrInstance, ArrInstanceInput, ArrOverview } from '../../domain/arr';
import type { ArrGateway } from '../../usecase/arrUseCases';

/** ArrGateway backed by the REST API. */
export class ArrApi implements ArrGateway {
  constructor(private readonly http: HttpClient) {}

  async instances(): Promise<ArrInstance[]> {
    return (await this.http.get<InstanceDTO[]>('/api/arr/instances')).map((d) => ({ ...d }));
  }
  async add(input: ArrInstanceInput): Promise<ArrInstance> {
    const body: InstanceInputDTO = input;
    return { ...(await this.http.post<InstanceDTO>('/api/arr/instances', body)) };
  }
  async update(id: string, input: ArrInstanceInput): Promise<ArrInstance> {
    const body: InstanceInputDTO = input;
    return { ...(await this.http.put<InstanceDTO>(`/api/arr/instances/${encodeURIComponent(id)}`, body)) };
  }
  delete(id: string): Promise<void> {
    return this.http.delete(`/api/arr/instances/${encodeURIComponent(id)}`);
  }
  async overview(): Promise<ArrOverview> {
    const d = await this.http.get<OverviewDTO>('/api/arr/overview');
    return {
      upcoming: d.upcoming.map((r) => ({ ...r, date: new Date(r.date) })),
      queue: d.queue,
      health: d.health,
      missing: d.missing,
      errors: d.errors,
    };
  }
}
