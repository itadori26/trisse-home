import type { HttpClient } from '../../../../shared/api-client/http';
import type { DiskDTO } from '../../../../shared/api-client/generated/diskhealth';
import type { Disk } from '../../domain/disk';
import type { DiskHealthGateway } from '../../usecase/diskHealthUseCases';

/** DiskHealthGateway backed by the REST API. */
export class DiskHealthApi implements DiskHealthGateway {
  constructor(private readonly http: HttpClient) {}

  async disks(): Promise<Disk[]> {
    return (await this.http.get<DiskDTO[]>('/api/diskhealth')).map((d) => ({
      ...d,
      checkedAt: new Date(d.checkedAt),
      temperature: d.temperature ?? null,
      powerOnHours: d.powerOnHours ?? null,
      reallocated: d.reallocated ?? null,
      pending: d.pending ?? null,
      percentageUsed: d.percentageUsed ?? null,
    }));
  }
}
