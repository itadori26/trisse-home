import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/immich';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/immich-requests';
import type { ImmichOverview, ImmichSettings } from '../../domain/immich';
import type { ImmichGateway } from '../../usecase/immichUseCases';

/** ImmichGateway backed by the REST API. */
export class ImmichApi implements ImmichGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<ImmichSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/immich/settings')) };
  }
  async saveSettings(url: string, apiKey: string): Promise<ImmichSettings> {
    const body: SettingsInputDTO = { url, apiKey };
    return { ...(await this.http.put<SettingsDTO>('/api/immich/settings', body)) };
  }
  async overview(): Promise<ImmichOverview> {
    return { ...(await this.http.get<OverviewDTO>('/api/immich/overview')) };
  }
}
