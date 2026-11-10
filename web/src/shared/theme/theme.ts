/** Supported color scheme preferences. */
export type ThemeMode = 'system' | 'light' | 'dark';

/** Resolves the effective scheme for a preference and the OS setting. */
export function resolveTheme(mode: ThemeMode, prefersDark: boolean): 'light' | 'dark' {
  if (mode === 'system') return prefersDark ? 'dark' : 'light';
  return mode;
}

const darkQuery = '(prefers-color-scheme: dark)';

/**
 * Applies the theme to the document root and follows OS changes while the
 * mode is "system". Returns a function that stops following.
 */
export function applyTheme(mode: ThemeMode): () => void {
  const media = window.matchMedia(darkQuery);
  const apply = () => {
    document.documentElement.dataset['theme'] = resolveTheme(mode, media.matches);
  };
  apply();
  media.addEventListener('change', apply);
  return () => media.removeEventListener('change', apply);
}
