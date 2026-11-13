import { fold } from './smart/text';

/**
 * Frecency of search choices (#159): which result was chosen for which
 * query, weighted by frequency and recency (as in Firefox and Alfred).
 * Each entry keeps one value that halves every HALF_LIFE_MS, so recording
 * a choice is "decay to now, add one" and no timestamps pile up.
 */

export interface FrecencyEntry {
  /** Normalised query prefix the choice was made for. */
  key: string;
  /** Id of the chosen result, e.g. "bookmark:12" or "answer:system.cpu". */
  target: string;
  /** Value at time `at`. */
  value: number;
  /** Milliseconds since the epoch. */
  at: number;
}

export const HALF_LIFE_MS = 14 * 24 * 3600 * 1000;
export const MAX_ENTRIES = 200;
/** Longest stored query prefix; longer queries are cut. */
export const MAX_KEY = 24;
/** Score multiplier per unit of boost, and the boost cap. */
export const BOOST_FACTOR = 0.5;
export const MAX_BOOST = 10;
/** Entries below this value are forgotten. */
const MIN_VALUE = 0.05;

/** The query as stored: folded, single spaces, cut to MAX_KEY. */
export function normaliseQuery(query: string): string {
  return fold(query).trim().replace(/\s+/g, ' ').slice(0, MAX_KEY);
}

/** Value of an entry at `now`, after exponential decay. */
export function decayed(entry: FrecencyEntry, now: number): number {
  const age = Math.max(0, now - entry.at);
  return entry.value * 0.5 ** (age / HALF_LIFE_MS);
}

/**
 * The boost of a target for a query: the decayed values of the choices
 * whose query starts with what is typed now, or that what is typed now
 * starts with ("pl" and "plex server" both profit from "plex").
 */
export function boost(entries: readonly FrecencyEntry[], query: string, target: string, now: number): number {
  const q = normaliseQuery(query);
  if (!q) return 0;
  let sum = 0;
  for (const e of entries)
    if (e.target === target && (e.key.startsWith(q) || q.startsWith(e.key))) sum += decayed(e, now);
  return Math.min(sum, MAX_BOOST);
}

/** A match score with the boost applied. */
export function boosted(score: number, b: number): number {
  return score * (1 + BOOST_FACTOR * b);
}

/**
 * Records a choice and returns the new entry list: decays the existing
 * entry and adds one, drops faded entries and evicts the least valuable
 * ones beyond MAX_ENTRIES.
 */
export function record(
  entries: readonly FrecencyEntry[],
  query: string,
  target: string,
  now: number,
): FrecencyEntry[] {
  const key = normaliseQuery(query);
  if (!key || !target) return [...entries];
  const old = entries.find((e) => e.key === key && e.target === target);
  const updated: FrecencyEntry = { key, target, value: (old ? decayed(old, now) : 0) + 1, at: now };
  const rest = entries.filter((e) => e !== old);
  return prune([updated, ...rest], now);
}

/** Drops faded entries and keeps the MAX_ENTRIES most valuable ones. */
export function prune(entries: readonly FrecencyEntry[], now: number): FrecencyEntry[] {
  return entries
    .map((e) => ({ e, v: decayed(e, now) }))
    .filter((x) => x.v >= MIN_VALUE)
    .sort((a, b) => b.v - a.v)
    .slice(0, MAX_ENTRIES)
    .map((x) => x.e);
}

/** Validates stored entries (storage may hold anything). */
export function parseEntries(data: unknown): FrecencyEntry[] {
  if (!Array.isArray(data)) return [];
  return data.filter(
    (e): e is FrecencyEntry =>
      typeof e === 'object' &&
      e !== null &&
      typeof e.key === 'string' &&
      typeof e.target === 'string' &&
      typeof e.value === 'number' &&
      Number.isFinite(e.value) &&
      typeof e.at === 'number' &&
      Number.isFinite(e.at),
  );
}
