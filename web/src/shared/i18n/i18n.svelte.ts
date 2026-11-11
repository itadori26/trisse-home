import { de } from './locales/de';
import { en, type MessageKey } from './locales/en';
import { formatMessage, type Locale, type MessageParams, resolveLocale } from './format';

const catalogs: Record<Locale, Record<MessageKey, string>> = { en, de };

const browserLanguages = (): readonly string[] =>
  typeof navigator === 'undefined' ? [] : (navigator.languages ?? [navigator.language]);

const state = $state<{ locale: Locale }>({ locale: resolveLocale('', browserLanguages()) });

/**
 * Translates a message key. Reading the locale here makes every template
 * that calls t() re-render when the language changes.
 */
export function t(key: MessageKey, params?: MessageParams): string {
  return formatMessage(catalogs[state.locale][key] ?? en[key], params);
}

/** The active UI language. */
export function locale(): Locale {
  return state.locale;
}

/** BCP 47 tag for Intl/toLocale* formatting in the active language. */
export function intlLocale(): string {
  return state.locale === 'de' ? 'de-DE' : (browserLanguages().find((l) => l.startsWith('en')) ?? 'en');
}

/** Applies a language preference ("" = browser language). */
export function setLanguage(preference: string): void {
  state.locale = resolveLocale(preference, browserLanguages());
  if (typeof document !== 'undefined') document.documentElement.lang = state.locale;
}

export type { MessageKey };
