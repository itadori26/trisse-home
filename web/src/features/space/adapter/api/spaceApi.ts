import type { HttpClient } from '../../../../shared/api-client/http';
import type { LaunchDTO, PictureDTO, SettingsDTO } from '../../../../shared/api-client/generated/space';
import type { Launch, LaunchStatus, Picture, SpaceSettings } from '../../domain/space';
import type { SpaceGateway } from '../../usecase/spaceUseCases';

/** SpaceGateway backed by the REST API. */
export class SpaceApi implements SpaceGateway {
  constructor(private readonly http: HttpClient) {}

  async launches(limit: number): Promise<Launch[]> {
    const d = await this.http.get<LaunchDTO[]>(`/api/space/launches?limit=${limit}`);
    return d.map((l) => ({ ...l, net: new Date(l.net), status: l.status as LaunchStatus }));
  }
  async picture(): Promise<Picture> {
    return { ...(await this.http.get<PictureDTO>('/api/space/apod')) };
  }
  async settings(): Promise<SpaceSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/space/settings')) };
  }
  async saveSettings(launches: boolean, picture: boolean, nasaKey: string): Promise<SpaceSettings> {
    return {
      ...(await this.http.put<SettingsDTO>('/api/space/settings', {
        launches,
        picture,
        nasaKey,
        keepKey: nasaKey === '',
      })),
    };
  }
}
