import type { Answer } from '../features/search/domain/smart/answer';
import type { SmartSearch } from '../features/search/usecase/smartSearch';
import type { Container } from './container';
import { searchUsage } from './searchUsage';

const engines = new WeakMap<Container, Promise<SmartSearch>>();

/**
 * Instant answers for the search widget (#156). The engine is loaded on the
 * first keystroke (not part of the initial bundle) and shared per session.
 * The widget offers no write actions; those stay in the quick search.
 */
export function widgetAnswers(c: Container): (query: string, signal: AbortSignal) => Promise<Answer[]> {
  return async (query, signal) => {
    let engine = engines.get(c);
    if (!engine) {
      engine = Promise.all([import('./searchSources'), c.enabledModules()]).then(([m, info]) =>
        m.createSmartSearch(c, info.modules, {
          canAct: false,
          weatherLocation: () => null,
          score: (q, target, score) => searchUsage.score(q, target, score),
        }),
      );
      engines.set(c, engine);
    }
    return (await engine).answer(query, signal);
  };
}
