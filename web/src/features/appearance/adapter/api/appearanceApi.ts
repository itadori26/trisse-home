import type { HttpClient } from '../../../../shared/api-client/http';
import type { AppearanceDTO } from '../../../../shared/api-client/generated/appearance';
import type { Appearance } from '../../domain/appearance';
import type { AppearanceGateway } from '../../usecase/appearanceUseCases';

const toAppearance = (d: AppearanceDTO): Appearance => ({
  theme: d.theme,
  accentColor: d.accentColor,
  backgroundImage: d.backgroundImage,
  customCss: d.customCss,
});

/** AppearanceGateway backed by the REST API. */
export class AppearanceApi implements AppearanceGateway {
  constructor(private readonly http: HttpClient) {}

  async load(): Promise<Appearance> {
    return toAppearance(await this.http.get<AppearanceDTO>('/api/appearance'));
  }

  async save(a: Appearance): Promise<Appearance> {
    const body: AppearanceDTO = { ...a };
    return toAppearance(await this.http.put<AppearanceDTO>('/api/appearance', body));
  }
}
