/**
 * Building blocks of the smart search answers (#161): every module adds its
 * intents and its answer source from its own src/app/search/<name>.ts, so
 * neither the intent catalogue nor searchSources.ts needs an edit (#193).
 */
import type { IntentSpec } from '../features/search/domain/smart/intent';
import type { SearchSource } from '../features/search/usecase/smartSearch';
import type { Container } from './container';

/** First widget config of a kind on the dashboard (e.g. the transit stop). */
export type WidgetConfig = (kind: string) => Record<string, unknown> | null;

/** What a source may need besides the container. */
export interface SearchContext {
  widgetConfig: WidgetConfig;
}

/** The search contribution of one module. */
export interface SearchContribution {
  /** Registered only while this backend module is enabled. */
  module: string;
  /** Intents as data; their `source` must be the id of `source()`. */
  intents: IntentSpec[];
  source: (container: Container, context: SearchContext) => SearchSource;
}

/** Identity helper for type checking in the module files. */
export function defineSearch(def: SearchContribution): SearchContribution {
  return def;
}

/** Helper for weighted terms: t(3, 'muell', 'abfall'). */
export const terms = (weight: number, ...words: string[]) => words.map((term) => ({ term, weight }));

/** Longest list in an answer. */
export const MAX_ITEMS = 6;
