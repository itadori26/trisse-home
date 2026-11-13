import type { ThemeMode } from '../../../shared/theme/theme';

/** Visual preferences of the start page. */
export interface Appearance {
  theme: ThemeMode;
  accentColor: string;
  backgroundImage: string;
  customCss: string;
  /** UI language; "" follows the browser. */
  language: '' | 'en' | 'de';
  /** Name in the greeting of the header (#189); may be empty. */
  greetingName: string;
  /** What the launcher lists (#189). */
  launcher: 'both' | 'services' | 'bookmarks';
}

/** Defaults before the preferences are loaded. */
export const defaultAppearance: Appearance = {
  theme: 'system',
  accentColor: '',
  backgroundImage: '',
  customCss: '',
  language: '',
  greetingName: '',
  launcher: 'both',
};

/**
 * CSS custom properties derived from the preferences. Empty values are
 * omitted so the theme defaults apply.
 */
export function cssVariables(a: Appearance): Record<string, string> {
  const vars: Record<string, string> = {};
  if (/^#[0-9a-f]{6}$/i.test(a.accentColor)) vars['--color-accent'] = a.accentColor;
  if (/^https?:\/\/[^\s"'()\\]+$/i.test(a.backgroundImage)) {
    vars['--background-image'] = `url("${a.backgroundImage}")`;
    // Slightly translucent cards keep the background visible.
    vars['--widget-opacity'] = '0.94';
  }
  return vars;
}
