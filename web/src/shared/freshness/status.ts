/** Data freshness of a widget (#94). */
export type Freshness = 'loading' | 'fresh' | 'stale' | 'error';

/** What a widget knows about its data. */
export interface WidgetStatus {
  /** Time of the last successful refresh (ms), null before the first one. */
  fetchedAt: number | null;
  /** Message of the last failed refresh, null after a success. */
  lastError: string | null;
  /** Refresh interval of the widget (ms). */
  intervalMs: number;
}

/** Data counts as stale after this many refresh intervals without success. */
export const STALE_FACTOR = 2;

/**
 * Classifies a status: an error wins (the last good value stays visible),
 * otherwise data older than STALE_FACTOR × interval is stale.
 */
export function freshness(s: WidgetStatus, now: number, factor = STALE_FACTOR): Freshness {
  if (s.lastError !== null) return 'error';
  if (s.fetchedAt === null) return 'loading';
  return now - s.fetchedAt > factor * s.intervalMs ? 'stale' : 'fresh';
}

const rank: Record<Freshness, number> = { fresh: 0, loading: 1, stale: 2, error: 3 };

/** Combines the statuses of a widget's jobs: the worst one counts, the oldest data is shown. */
export function combine(
  list: readonly WidgetStatus[],
  now: number,
): { freshness: Freshness; fetchedAt: number | null; lastError: string | null } {
  let worst: Freshness = 'fresh';
  let fetchedAt: number | null = null;
  let lastError: string | null = null;
  for (const s of list) {
    const f = freshness(s, now);
    if (rank[f] > rank[worst]) worst = f;
    if (s.fetchedAt !== null && (fetchedAt === null || s.fetchedAt < fetchedAt)) fetchedAt = s.fetchedAt;
    lastError ??= s.lastError;
  }
  return { freshness: worst, fetchedAt, lastError };
}

/** Whole minutes since a timestamp (for "updated x min ago"). */
export function minutesAgo(at: number, now: number): number {
  return Math.max(0, Math.floor((now - at) / 60_000));
}
