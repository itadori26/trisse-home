import type { HttpClient } from '../../../../shared/api-client/http';
import type { SettingsDTO, StatsDTO } from '../../../../shared/api-client/generated/adguard';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/adguard-requests';
import type { AdGuardSettings, AdGuardStats, DnsProvider } from '../../domain/adguard';
import type { AdGuardGateway } from '../../usecase/adguardUseCases';

const query = (p: DnsProvider) => `?provider=${encodeURIComponent(p)}`;

/** AdGuardGateway backed by the REST API. */
export class AdGuardApi implements AdGuardGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(provider: DnsProvider): Promise<AdGuardSettings> {
    return { ...(await this.http.get<SettingsDTO>(`/api/adguard/settings${query(provider)}`)) };
  }
  async saveSettings(
    provider: DnsProvider,
    url: string,
    username: string,
    password: string,
  ): Promise<AdGuardSettings> {
    const body: SettingsInputDTO = { url, username, password };
    return { ...(await this.http.put<SettingsDTO>(`/api/adguard/settings${query(provider)}`, body)) };
  }
  async stats(provider: DnsProvider): Promise<AdGuardStats> {
    const d = await this.http.get<StatsDTO>(`/api/adguard/stats${query(provider)}`);
    return {
      queries: d.queries,
      blocked: d.blocked,
      blockedPercent: d.blockedPercent,
      avgProcessingMs: d.avgProcessingMs ?? null,
      protectionEnabled: d.protectionEnabled,
      topBlocked: d.topBlocked,
      error: d.error,
    };
  }
}
