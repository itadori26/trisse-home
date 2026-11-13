/**
 * German texts of all features, collected from features/<name>/i18n/de.ts
 * (#193). A separate chunk from the English one, so only the active language
 * is downloaded.
 */
const parts = import.meta.glob<{ de: Record<string, string> }>('../features/*/i18n/de.ts', {
  eager: true,
});

export const messages: Record<string, string> = Object.assign({}, ...Object.values(parts).map((m) => m.de));
