import type { HttpClient } from '../../../../shared/api-client/http';
import type { SnapshotDTO } from '../../../../shared/api-client/generated/system';
import type { SystemSnapshot } from '../../domain/system';
import type { SystemGateway } from '../../usecase/systemUseCases';

/** SystemGateway backed by the REST API. */
export class SystemApi implements SystemGateway {
  constructor(private readonly http: HttpClient) {}

  async snapshot(): Promise<SystemSnapshot> {
    const d = await this.http.get<SnapshotDTO>('/api/system');
    return {
      cpuPercent: d.cpuPercent,
      cpuCount: d.cpuCount,
      load: [d.load[0] ?? 0, d.load[1] ?? 0, d.load[2] ?? 0],
      memory: d.memory,
      swap: d.swap,
      disks: d.disks,
      uptimeSeconds: d.uptimeSeconds,
      temperature: d.temperature,
    };
  }
}
