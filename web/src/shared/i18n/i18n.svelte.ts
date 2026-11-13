import type { MessageKey } from './locales/en';
import { formatMessage, type Locale, type MessageParams, resolveLocale } from './format';
import type { FeatureLoaders } from './registry';

type Catalogue = Record<MessageKey, string>;

/**
 * Both catalogues are loaded on demand, so only the active language is
 * downloaded and neither is part of the initial bundle (ADR 0011, #116).
 * App.svelte awaits loadLanguage() before the first render.
 */
let english = $state.raw<Catalogue | null>(null);
let german = $state.raw<Catalogue | null>(null);

/** Loaders of the feature catalogues, registered by src/app (#193). */
let featureLoaders: FeatureLoaders | undefined;

/**
 * Registers where the texts of the features come from. They are merged into
 * the shared catalogue when a language loads, so features add their texts
 * without editing a shared file.
 */
export function registerFeatureMessages(loaders: FeatureLoaders): void {
  featureLoaders = loaders;
}

const browserLanguages = (): readonly string[] =>
  typeof navigator === 'undefined' ? [] : (navigator.languages ?? [navigator.language]);

const state = $state<{ locale: Locale }>({ locale: resolveLocale('', browserLanguages()) });

/**
 * Translates a message key. Reading the locale here makes every template
 * that calls t() re-render when the language changes.
 */
export function t(key: MessageKey, params?: MessageParams): string {
  // While a switched-to language loads, the other catalogue stays visible.
  const catalog = state.locale === 'de' ? (german ?? english) : (english ?? german);
  return formatMessage(catalog?.[key] ?? english?.[key] ?? key, params);
}

/** Loads the catalogues of the active language; resolves when they are ready. */
export async function loadLanguage(): Promise<void> {
  if (state.locale === 'de') {
    german ??= await merged((await import('./locales/de')).de, featureLoaders?.de);
  } else {
    english ??= await merged((await import('./locales/en')).en, featureLoaders?.en);
  }
  await loadServerMessages();
}

async function merged(
  shared: object,
  features: (() => Promise<Record<string, string>>) | undefined,
): Promise<Catalogue> {
  return { ...shared, ...(await features?.()) } as Catalogue;
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
  void loadLanguage();
}

export type { MessageKey };

/**
 * Translates a backend error message (#102). Returns undefined for unknown
 * keys, so callers fall back to the English text sent by the server.
 */
export function serverMessage(key: string | undefined, params?: Record<string, string>): string | undefined {
  // The server already answers in English; translations are loaded lazily.
  if (!key || state.locale !== 'de' || !serverCatalogue.de) return undefined;
  const template = serverCatalogue.de[key];
  return template === undefined ? undefined : formatMessage(template, params);
}

/** Lazily loaded backend message translations (not in the initial bundle, ADR 0011). */
const serverCatalogue: { de?: Record<string, string> } = {};

/** Loads the backend message translations of the active language. */
export async function loadServerMessages(): Promise<void> {
  if (state.locale === 'de' && !serverCatalogue.de) {
    serverCatalogue.de = (await import('./locales/server-de')).serverDe;
  }
}
