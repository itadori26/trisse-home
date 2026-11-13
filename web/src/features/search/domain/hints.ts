/**
 * Rotating placeholder of the search pill (#189): questions and ideas that
 * show what the search can do. Pure rules; the timer lives in the UI.
 */

/** Message keys of the hints, in display order. */
export const SEARCH_HINTS = [
  'search.hint.learn',
  'search.hint.todo',
  'search.hint.cpu',
  'search.hint.weather',
  'search.hint.note',
  'search.hint.bang',
] as const;

export type SearchHint = (typeof SEARCH_HINTS)[number];

/** How long one hint is shown. */
export const HINT_INTERVAL_MS = 4000;

/**
 * The hint to show: rotation pauses while the field has focus (the current
 * hint stays) and is off for reduced motion (always the first hint).
 */
export function nextHint(current: number, opts: { focused: boolean; reducedMotion: boolean }): number {
  if (opts.reducedMotion) return 0;
  if (opts.focused) return current;
  return (current + 1) % SEARCH_HINTS.length;
}
