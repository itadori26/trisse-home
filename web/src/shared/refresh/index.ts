import { getContext, setContext } from 'svelte';
import { browserClock, browserTicker, documentVisibility } from './browser';
import { type RefreshHandle, RefreshScheduler } from './scheduler';

export { RefreshScheduler, type RefreshHandle } from './scheduler';
export { observeViewport } from './browser';

/** Creates the app's scheduler with the browser adapters (composition root). */
export function createRefreshScheduler(): RefreshScheduler {
  return new RefreshScheduler(browserClock, browserTicker, documentVisibility);
}

/** What a widget's surroundings provide: the scheduler and its viewport state. */
export interface RefreshContext {
  scheduler: RefreshScheduler;
  /** Registers a handle so the surrounding card can report viewport changes. */
  track(handle: RefreshHandle): () => void;
}

const key = Symbol('refresh');

/** Provides the refresh context to the components below (app layer). */
export function provideRefresh(ctx: RefreshContext): void {
  setContext(key, ctx);
}

let fallback: RefreshScheduler | undefined;

/**
 * Returns poll(run, intervalMs) for a widget; call it during component
 * initialisation and use the result inside an $effect:
 *
 *     const poll = usePolling();
 *     $effect(() => poll(() => api.load().then(…), 60_000));
 *
 * run is called at once and then by the shared scheduler (#117); a rejected
 * promise counts as a failure and backs off. The returned function stops
 * polling, so it can be returned from the $effect.
 */
export function usePolling(): (run: () => Promise<unknown>, intervalMs: number) => () => void {
  const ctx = getContext<RefreshContext | undefined>(key);
  const scheduler = ctx?.scheduler ?? (fallback ??= createRefreshScheduler());
  return (run, intervalMs) => {
    const handle = scheduler.register(run, intervalMs);
    const untrack = ctx?.track(handle) ?? (() => {});
    return () => {
      untrack();
      handle.dispose();
    };
  };
}
