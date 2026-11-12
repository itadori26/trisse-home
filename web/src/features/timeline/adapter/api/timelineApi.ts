import type { HttpClient } from '../../../../shared/api-client/http';
import type { EventDTO, SettingsDTO } from '../../../../shared/api-client/generated/timeline';
import type { SettingsDTO as SettingsInputDTO } from '../../../../shared/api-client/generated/timeline-requests';
import type { TimelineEvent } from '../../domain/timeline';
import type { TimelineFilter, TimelineGateway } from '../../usecase/timelineUseCases';

/** TimelineGateway backed by the REST API. */
export class TimelineApi implements TimelineGateway {
  constructor(private readonly http: HttpClient) {}

  async list(f: TimelineFilter): Promise<TimelineEvent[]> {
    const q = new URLSearchParams({ source: f.source, severity: f.severity, limit: String(f.limit) });
    return (await this.http.get<EventDTO[]>(`/api/timeline?${q}`)).map((d) => ({ ...d, at: new Date(d.at) }));
  }

  async retentionDays(): Promise<number> {
    return (await this.http.get<SettingsDTO>('/api/timeline/settings')).retentionDays;
  }

  async setRetentionDays(days: number): Promise<number> {
    const body: SettingsInputDTO = { retentionDays: days };
    return (await this.http.put<SettingsDTO>('/api/timeline/settings', body)).retentionDays;
  }
}
