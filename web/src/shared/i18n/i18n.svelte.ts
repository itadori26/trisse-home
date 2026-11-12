import { de } from './locales/de';
import { serverDe } from './locales/server-de';
import { serverMessages } from '../api-client/generated/server-messages';
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

/**
 * Translates a backend error message (#102). Returns undefined for unknown
 * keys, so callers fall back to the English text sent by the server.
 */
export function serverMessage(key: string | undefined, params?: Record<string, string>): string | undefined {
  if (!key) return undefined;
  const catalogue: Record<string, string> = state.locale === 'de' ? serverDe : serverMessages;
  const template = catalogue[key];
  return template === undefined ? undefined : formatMessage(template, params);
}
