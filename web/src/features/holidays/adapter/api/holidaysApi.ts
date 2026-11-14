import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/holidays';
import type { HolidaysOverview, State } from '../../domain/holidays';
import type { HolidaysGateway } from '../../usecase/holidaysUseCases';

/** Parses YYYY-MM-DD as a local date. */
const localDate = (s: string) => {
  const [y, m, d] = s.split('-').map(Number);
  return new Date(y ?? 1970, (m ?? 1) - 1, d ?? 1);
};

/** HolidaysGateway backed by the REST API. */
export class HolidaysApi implements HolidaysGateway {
  constructor(private readonly http: HttpClient) {}

  async upcoming(state: State, limit: number): Promise<HolidaysOverview> {
    const q = new URLSearchParams({ state, limit: String(limit) });
    const d = await this.http.get<OverviewDTO>(`/api/holidays?${q}`);
    return {
      state: d.state as State,
      school: d.school,
      schoolError: d.schoolError,
      schoolUpdatedAt: d.schoolUpdatedAt ? new Date(d.schoolUpdatedAt) : null,
      entries: d.entries.map((e) => ({
        ...e,
        kind: e.kind === 'school' ? 'school' : 'public',
        start: localDate(e.start),
        end: localDate(e.end),
      })),
    };
  }
  async settings(): Promise<{ schoolUrl: string }> {
    return { ...(await this.http.get<SettingsDTO>('/api/holidays/settings')) };
  }
  async saveSettings(schoolUrl: string): Promise<{ schoolUrl: string }> {
    return { ...(await this.http.put<SettingsDTO>('/api/holidays/settings', { schoolUrl })) };
  }
}
