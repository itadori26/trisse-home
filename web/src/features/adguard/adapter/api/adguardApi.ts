import type { HttpClient } from '../../../../shared/api-client/http';
import type { SettingsDTO, StatsDTO } from '../../../../shared/api-client/generated/adguard';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/adguard-requests';
import type { AdGuardSettings, AdGuardStats } from '../../domain/adguard';
import type { AdGuardGateway } from '../../usecase/adguardUseCases';

/** AdGuardGateway backed by the REST API. */
export class AdGuardApi implements AdGuardGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<AdGuardSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/adguard/settings')) };
  }
  async saveSettings(url: string, username: string, password: string): Promise<AdGuardSettings> {
    const body: SettingsInputDTO = { url, username, password };
    return { ...(await this.http.put<SettingsDTO>('/api/adguard/settings', body)) };
  }
  async stats(): Promise<AdGuardStats> {
    const d = await this.http.get<StatsDTO>('/api/adguard/stats');
    return {
      queries: d.queries,
      blocked: d.blocked,
      blockedPercent: d.blockedPercent,
      avgProcessingMs: d.avgProcessingMs,
      protectionEnabled: d.protectionEnabled,
      topBlocked: d.topBlocked,
      error: d.error,
    };
  }
}
