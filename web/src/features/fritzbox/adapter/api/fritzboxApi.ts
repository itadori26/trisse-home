import type { HttpClient } from '../../../../shared/api-client/http';
import type { SettingsDTO, StatusDTO } from '../../../../shared/api-client/generated/fritzbox';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/fritzbox-requests';
import type { FritzBoxSettings, FritzBoxStatus } from '../../domain/fritzbox';
import type { FritzBoxGateway } from '../../usecase/fritzboxUseCases';

/** FritzBoxGateway backed by the REST API. */
export class FritzBoxApi implements FritzBoxGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<FritzBoxSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/fritzbox/settings')) };
  }
  async saveSettings(url: string, username: string, password: string): Promise<FritzBoxSettings> {
    const body: SettingsInputDTO = { url, username, password };
    return { ...(await this.http.put<SettingsDTO>('/api/fritzbox/settings', body)) };
  }
  async status(): Promise<FritzBoxStatus> {
    const d = await this.http.get<StatusDTO>('/api/fritzbox/status');
    return {
      model: d.model,
      wan: { ...d.wan },
      calls: d.calls.map((c) => ({ ...c, at: new Date(c.at) })),
      hint: d.hint,
      error: d.error,
    };
  }
}
