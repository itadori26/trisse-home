/**
 * English texts of all features, collected from features/<name>/i18n/en.ts
 * (#193). Loaded as one chunk when the English catalogue loads, so a new
 * feature adds its texts without touching this file.
 */
const parts = import.meta.glob<{ en: Record<string, string> }>('../features/*/i18n/en.ts', {
  eager: true,
});

export const messages: Record<string, string> = Object.assign({}, ...Object.values(parts).map((m) => m.en));
