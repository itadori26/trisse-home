import type { Clock, Ticker, VisibilitySource } from './scheduler';

/** Clock backed by Date.now. */
export const browserClock: Clock = { now: () => Date.now() };

/** Ticker backed by setInterval. */
export const browserTicker: Ticker = {
  every(ms, fn) {
    const id = setInterval(fn, ms);
    return () => clearInterval(id);
  },
};

/** Page visibility from the Page Visibility API (always visible outside a browser). */
export const documentVisibility: VisibilitySource = {
  visible: () => typeof document === 'undefined' || document.visibilityState !== 'hidden',
  onChange(fn) {
    if (typeof document === 'undefined') return () => {};
    document.addEventListener('visibilitychange', fn);
    return () => document.removeEventListener('visibilitychange', fn);
  },
};

/**
 * Observes whether an element is in the viewport; calls fn with the
 * current state. Without IntersectionObserver it reports "visible".
 */
export function observeViewport(el: Element, fn: (visible: boolean) => void): () => void {
  if (typeof IntersectionObserver === 'undefined') {
    fn(true);
    return () => {};
  }
  const io = new IntersectionObserver((entries) => {
    for (const e of entries) fn(e.isIntersecting);
  });
  io.observe(el);
  return () => io.disconnect();
}
