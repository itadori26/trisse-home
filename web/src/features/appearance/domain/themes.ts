/** A theme of the gallery (#111): design tokens only. */
export interface ThemeInfo {
  id: string;
  name: string;
  author: string;
  scheme: 'light' | 'dark';
  tokens: Record<string, string>;
  builtIn: boolean;
}

/** The theme gallery. */
export interface ThemeGallery {
  themes: ThemeInfo[];
  /** Id of the applied theme, "" for own or no custom CSS. */
  applied: string;
  canRestore: boolean;
}

/** Largest theme file accepted for import. */
export const MAX_THEME_BYTES = 16 << 10;

const safeValue = /^(#[0-9a-f]{3}|#[0-9a-f]{6}|\d{1,2}px)$/i;

/**
 * Inline style that previews a theme on its card: the tokens as custom
 * properties, scoped to the card. Values are checked again here even though
 * the server validated them, so nothing but colors and radii gets through.
 */
export function previewStyle(theme: Pick<ThemeInfo, 'tokens'>): string {
  return Object.entries(theme.tokens)
    .filter(([k, v]) => /^--[a-z0-9-]+$/.test(k) && safeValue.test(v))
    .map(([k, v]) => `${k}: ${v}`)
    .join('; ');
}
