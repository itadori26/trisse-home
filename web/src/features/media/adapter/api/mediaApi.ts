import type { HttpClient } from '../../../../shared/api-client/http';
import type { OverviewDTO, SettingsDTO } from '../../../../shared/api-client/generated/media';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/media-requests';
import type { MediaOverview, MediaProvider, MediaSettings } from '../../domain/media';
import type { MediaGateway } from '../../usecase/mediaUseCases';

/** MediaGateway backed by the REST API. */
export class MediaApi implements MediaGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<MediaSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/media/settings')) };
  }
  async saveSettings(provider: MediaProvider, url: string, token: string): Promise<MediaSettings> {
    const body: SettingsInputDTO = { provider, url, token };
    return { ...(await this.http.put<SettingsDTO>('/api/media/settings', body)) };
  }
  async overview(): Promise<MediaOverview> {
    const d = await this.http.get<OverviewDTO>('/api/media/overview');
    return { serverName: d.serverName, recent: d.recent, sessions: d.sessions };
  }
}
