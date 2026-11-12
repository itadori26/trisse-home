import type { HttpClient } from '../../../../shared/api-client/http';
import type { CreatedDTO, SourceDTO } from '../../../../shared/api-client/generated/backupmonitor';
import type { SourceInputDTO } from '../../../../shared/api-client/generated/backupmonitor-requests';
import type { BackupSource, CreatedSource, SourceSchedule } from '../../domain/source';
import type { BackupMonitorGateway } from '../../usecase/backupMonitorUseCases';

const toSource = (d: SourceDTO): BackupSource => ({
  id: d.id,
  name: d.name,
  maxAgeHours: d.maxAgeHours,
  kind: d.kind,
  cron: d.cron,
  graceMinutes: d.graceMinutes,
  lastStartAt: d.lastStartAt ? new Date(d.lastStartAt) : null,
  lastDurationSeconds: d.lastDurationSeconds,
  state: d.state,
  lastSuccessAt: d.lastSuccessAt ? new Date(d.lastSuccessAt) : null,
  lastFailureAt: d.lastFailureAt ? new Date(d.lastFailureAt) : null,
  lastMessage: d.lastMessage,
});
const toCreated = (d: CreatedDTO): CreatedSource => ({ source: toSource(d.source), pingUrl: d.pingUrl });
const path = (id: string) => `/api/backup-status/sources/${encodeURIComponent(id)}`;

/** BackupMonitorGateway backed by the REST API. */
export class BackupMonitorApi implements BackupMonitorGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<BackupSource[]> {
    return (await this.http.get<SourceDTO[]>('/api/backup-status/sources')).map(toSource);
  }
  async create(name: string, schedule: SourceSchedule): Promise<CreatedSource> {
    const body: SourceInputDTO = { name, ...schedule };
    return toCreated(await this.http.post<CreatedDTO>('/api/backup-status/sources', body));
  }
  async update(id: string, name: string, schedule: SourceSchedule): Promise<BackupSource> {
    const body: SourceInputDTO = { name, ...schedule };
    return toSource(await this.http.put<SourceDTO>(path(id), body));
  }
  delete(id: string): Promise<void> {
    return this.http.delete(path(id));
  }
  async rotateToken(id: string): Promise<CreatedSource> {
    return toCreated(await this.http.post<CreatedDTO>(`${path(id)}/token`));
  }
}
