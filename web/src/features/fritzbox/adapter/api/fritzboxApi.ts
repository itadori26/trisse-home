import type { HttpClient } from '../../../../shared/api-client/http';
import type { CallsDTO, SettingsDTO, StatusDTO } from '../../../../shared/api-client/generated/fritzbox';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/fritzbox-requests';
import type { CallFilterKind, CallList, FritzBoxSettings, FritzBoxStatus } from '../../domain/fritzbox';
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
      hint: d.hint,
      error: d.error,
    };
  }
  async calls(filter: CallFilterKind, limit: number): Promise<CallList> {
    const q = new URLSearchParams({ filter, limit: String(limit) });
    const d = await this.http.get<CallsDTO>(`/api/fritzbox/calls?${q}`);
    return {
      calls: d.calls.map((c) => ({ ...c, at: new Date(c.at) })),
      hint: d.hint,
      error: d.error,
      updatedAt: d.updatedAt.startsWith('0001-') ? null : new Date(d.updatedAt),
    };
  }
}
