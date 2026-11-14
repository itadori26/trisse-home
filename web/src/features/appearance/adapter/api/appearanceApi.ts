import type { HttpClient } from '../../../../shared/api-client/http';
import type { AppearanceDTO, GalleryDTO, ThemeDTO } from '../../../../shared/api-client/generated/appearance';
import type { Appearance } from '../../domain/appearance';
import type { ThemeGallery, ThemeInfo } from '../../domain/themes';
import type { AppearanceGateway } from '../../usecase/appearanceUseCases';

const toAppearance = (d: AppearanceDTO): Appearance => ({
  theme: d.theme,
  accentColor: d.accentColor,
  backgroundImage: d.backgroundImage,
  customCss: d.customCss,
  language: d.language ?? '',
  greetingName: d.greetingName ?? '',
  launcher: d.launcher ?? 'both',
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

  gallery(): Promise<ThemeGallery> {
    return this.http.get<GalleryDTO>('/api/appearance/themes');
  }

  async applyTheme(id: string): Promise<Appearance> {
    return toAppearance(
      await this.http.post<AppearanceDTO>(`/api/appearance/themes/${encodeURIComponent(id)}/apply`),
    );
  }

  async restoreStyle(): Promise<Appearance> {
    return toAppearance(await this.http.post<AppearanceDTO>('/api/appearance/themes/restore'));
  }

  importTheme(json: string): Promise<ThemeInfo> {
    return this.http.postText<ThemeDTO>('/api/appearance/themes', json, 'application/json');
  }

  deleteTheme(id: string): Promise<void> {
    return this.http.delete(`/api/appearance/themes/${encodeURIComponent(id)}`);
  }

  exportUrl(id: string): string {
    return `/api/appearance/themes/${encodeURIComponent(id)}`;
  }
}
