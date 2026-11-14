import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/waste';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/waste-requests';
import type { WasteOverview, WasteRule, WasteSettings, WasteSource } from '../../domain/waste';
import type { WasteGateway } from '../../usecase/wasteUseCases';

/** Parses YYYY-MM-DD as a local date. */
const localDate = (s: string) => {
  const [y, m, d] = s.split('-').map(Number);
  return new Date(y ?? 1970, (m ?? 1) - 1, d ?? 1);
};

const toSettings = (d: SettingsDTO): WasteSettings => ({
  configured: d.configured,
  source: d.source as WasteSource,
  url: d.url,
  file: d.file
    ? {
        name: d.file.name,
        uploadedAt: new Date(d.file.uploadedAt),
        events: d.file.events,
        from: localDate(d.file.from),
        to: localDate(d.file.to),
      }
    : null,
  rules: d.rules.map((r) => ({ ...r })),
});

/** WasteGateway backed by the REST API. */
export class WasteApi implements WasteGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<WasteSettings> {
    return toSettings(await this.http.get<SettingsDTO>('/api/waste/settings'));
  }
  async saveSettings(url: string, rules: WasteRule[]): Promise<WasteSettings> {
    const body: SettingsInputDTO = { url, rules };
    return toSettings(await this.http.put<SettingsDTO>('/api/waste/settings', body));
  }
  async uploadCalendar(file: File): Promise<WasteSettings> {
    const form = new FormData();
    form.append('file', file, file.name);
    return toSettings(await this.http.postForm<SettingsDTO>('/api/waste/calendar', form));
  }
  async removeCalendar(): Promise<WasteSettings> {
    return toSettings(await this.http.delete<SettingsDTO>('/api/waste/calendar'));
  }
  async upcoming(): Promise<WasteOverview> {
    const d = await this.http.get<OverviewDTO>('/api/waste/upcoming');
    return {
      types: d.types.map((t) => ({ ...t, dates: (t.dates ?? []).map(localDate) })),
      error: d.error,
      endsOn: d.endsOn ? localDate(d.endsOn) : null,
    };
  }
}
