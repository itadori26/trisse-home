import type { HttpClient } from '../../../../shared/api-client/http';
import type { SettingsDTO, StatusDTO } from '../../../../shared/api-client/generated/nextcloud';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/nextcloud-requests';
import type { NextcloudSettings, NextcloudStatus } from '../../domain/nextcloud';
import type { NextcloudGateway } from '../../usecase/nextcloudUseCases';

/** NextcloudGateway backed by the REST API. */
export class NextcloudApi implements NextcloudGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<NextcloudSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/nextcloud/settings')) };
  }
  async saveSettings(url: string, username: string, appPassword: string): Promise<NextcloudSettings> {
    const body: SettingsInputDTO = { url, username, appPassword };
    return { ...(await this.http.put<SettingsDTO>('/api/nextcloud/settings', body)) };
  }
  async status(): Promise<NextcloudStatus> {
    const d = await this.http.get<StatusDTO>('/api/nextcloud/status');
    return {
      version: d.version,
      maintenance: d.maintenance,
      quota: { ...d.quota },
      server: d.server ? { ...d.server } : null,
      serverHint: d.serverHint,
      error: d.error,
    };
  }
}
