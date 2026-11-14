import { t } from '../../../shared/i18n';
import type { Appearance } from '../domain/appearance';
import { MAX_THEME_BYTES, type ThemeGallery, type ThemeInfo } from '../domain/themes';

/** Output port: appearance backend. */
export interface AppearanceGateway {
  load(): Promise<Appearance>;
  save(a: Appearance): Promise<Appearance>;
  gallery(): Promise<ThemeGallery>;
  applyTheme(id: string): Promise<Appearance>;
  restoreStyle(): Promise<Appearance>;
  importTheme(json: string): Promise<ThemeInfo>;
  deleteTheme(id: string): Promise<void>;
  exportUrl(id: string): string;
}

/** Application logic of the appearance feature. */
export class AppearanceUseCases {
  constructor(private readonly gateway: AppearanceGateway) {}

  load(): Promise<Appearance> {
    return this.gateway.load();
  }

  /** Saves trimmed preferences. */
  save(a: Appearance): Promise<Appearance> {
    return this.gateway.save({
      ...a,
      accentColor: a.accentColor.trim(),
      backgroundImage: a.backgroundImage.trim(),
    });
  }

  /** The theme gallery (#111). */
  gallery(): Promise<ThemeGallery> {
    return this.gateway.gallery();
  }

  /** Applies a theme; own custom CSS and accent color are kept for restoring. */
  applyTheme(id: string): Promise<Appearance> {
    return this.gateway.applyTheme(id);
  }

  /** Brings back the style a theme replaced. */
  restoreStyle(): Promise<Appearance> {
    return this.gateway.restoreStyle();
  }

  /** Imports a shared theme file. */
  async importTheme(file: { size: number; text(): Promise<string> }): Promise<ThemeInfo> {
    if (file.size > MAX_THEME_BYTES) throw new Error(t('appearance.theme_file_too_large'));
    return this.gateway.importTheme(await file.text());
  }

  deleteTheme(id: string): Promise<void> {
    return this.gateway.deleteTheme(id);
  }

  /** Download address of the shareable theme file. */
  exportUrl(id: string): string {
    return this.gateway.exportUrl(id);
  }
}
