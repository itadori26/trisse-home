import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/waste';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/waste-requests';
import type { WasteOverview, WasteRule, WasteSettings } from '../../domain/waste';
import type { WasteGateway } from '../../usecase/wasteUseCases';

/** Parses YYYY-MM-DD as a local date. */
const localDate = (s: string) => {
  const [y, m, d] = s.split('-').map(Number);
  return new Date(y ?? 1970, (m ?? 1) - 1, d ?? 1);
};

/** WasteGateway backed by the REST API. */
export class WasteApi implements WasteGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<WasteSettings> {
    const d = await this.http.get<SettingsDTO>('/api/waste/settings');
    return { configured: d.configured, url: d.url, rules: d.rules.map((r) => ({ ...r })) };
  }
  async saveSettings(url: string, rules: WasteRule[]): Promise<WasteSettings> {
    const body: SettingsInputDTO = { url, rules };
    const d = await this.http.put<SettingsDTO>('/api/waste/settings', body);
    return { configured: d.configured, url: d.url, rules: d.rules.map((r) => ({ ...r })) };
  }
  async upcoming(): Promise<WasteOverview> {
    const d = await this.http.get<OverviewDTO>('/api/waste/upcoming');
    return {
      types: d.types.map((t) => ({ ...t, dates: (t.dates ?? []).map(localDate) })),
      error: d.error,
    };
  }
}
