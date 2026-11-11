/** Values that can be inserted into a message. */
export type MessageParams = Record<string, string | number>;

/** Replaces {name} placeholders; unknown placeholders stay visible. */
export function formatMessage(template: string, params?: MessageParams): string {
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (m, name: string) => (name in params ? String(params[name]) : m));
}

/** Names of the placeholders in a message, sorted. */
export function placeholders(template: string): string[] {
  return [...template.matchAll(/\{(\w+)\}/g)].map((m) => m[1] ?? '').sort();
}

/** Supported UI languages. */
export type Locale = 'en' | 'de';

/** Picks the UI language: explicit preference, else the browser languages. */
export function resolveLocale(preference: string, browser: readonly string[]): Locale {
  if (preference === 'en' || preference === 'de') return preference;
  for (const l of browser) {
    const base = l.toLowerCase().split('-')[0];
    if (base === 'de' || base === 'en') return base;
  }
  return 'en';
}
