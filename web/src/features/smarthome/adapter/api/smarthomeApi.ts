import type { HttpClient } from '../../../../shared/api-client/http';
import type { EntityDTO, SettingsDTO } from '../../../../shared/api-client/generated/smarthome';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/smarthome-requests';
import type { Entity, SmartHomeSettings } from '../../domain/smarthome';
import type { SmartHomeGateway } from '../../usecase/smarthomeUseCases';

const ZERO_TIME = '0001-01-01T00:00:00Z';

const toEntity = (d: EntityDTO): Entity => ({
  ...d,
  lastChanged: d.lastChanged && d.lastChanged !== ZERO_TIME ? new Date(d.lastChanged) : null,
});

/** SmartHomeGateway backed by the REST API. */
export class SmarthomeApi implements SmartHomeGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<SmartHomeSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/smarthome/settings')) };
  }
  async saveSettings(url: string, token: string): Promise<SmartHomeSettings> {
    const body: SettingsInputDTO = { url, token };
    return { ...(await this.http.put<SettingsDTO>('/api/smarthome/settings', body)) };
  }
  async entities(): Promise<Entity[]> {
    return (await this.http.get<EntityDTO[]>('/api/smarthome/entities')).map(toEntity);
  }
  async states(ids: string[]): Promise<Entity[]> {
    const q = new URLSearchParams({ ids: ids.join(',') });
    return (await this.http.get<EntityDTO[]>(`/api/smarthome/states?${q}`)).map(toEntity);
  }
  act(id: string): Promise<void> {
    return this.http.post(`/api/smarthome/entities/${encodeURIComponent(id)}/action`);
  }
}
