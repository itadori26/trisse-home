import { describe, expect, it } from 'vitest';
import {
  boost,
  boosted,
  decayed,
  type FrecencyEntry,
  HALF_LIFE_MS,
  MAX_BOOST,
  MAX_ENTRIES,
  normaliseQuery,
  parseEntries,
  record,
} from './frecency';

const DAY = 24 * 3600 * 1000;
const T0 = Date.UTC(2026, 10, 13);

describe('frecency', () => {
  it('normalises the query', () => {
    expect(normaliseQuery('  Plex   Server ')).toBe('plex server');
    expect(normaliseQuery('Büro')).toBe('buro');
    expect(normaliseQuery('x'.repeat(40))).toHaveLength(24);
  });

  it('halves a choice every half-life', () => {
    const e: FrecencyEntry = { key: 'plex', target: 'service:1', value: 1, at: T0 };
    expect(decayed(e, T0)).toBe(1);
    expect(decayed(e, T0 + HALF_LIFE_MS)).toBeCloseTo(0.5);
    expect(decayed(e, T0 + 2 * HALF_LIFE_MS)).toBeCloseTo(0.25);
    expect(decayed(e, T0 - DAY)).toBe(1); // clock went backwards
  });

  it('adds up repeated choices and decays the old part first', () => {
    let entries = record([], 'plex', 'service:1', T0);
    entries = record(entries, 'Plex', 'service:1', T0 + HALF_LIFE_MS);
    expect(entries).toHaveLength(1);
    expect(entries[0]?.value).toBeCloseTo(1.5);
  });

  it('boosts by prefix in both directions, per target', () => {
    const entries = record([], 'plex', 'service:1', T0);
    expect(boost(entries, 'pl', 'service:1', T0)).toBe(1);
    expect(boost(entries, 'plex server', 'service:1', T0)).toBe(1);
    expect(boost(entries, 'plex', 'bookmark:9', T0)).toBe(0);
    expect(boost(entries, 'jelly', 'service:1', T0)).toBe(0);
    expect(boost(entries, '  ', 'service:1', T0)).toBe(0);
  });

  it('caps the boost', () => {
    let entries: FrecencyEntry[] = [];
    for (let i = 0; i < 50; i++) entries = record(entries, 'plex', 'service:1', T0);
    expect(boost(entries, 'plex', 'service:1', T0)).toBe(MAX_BOOST);
  });

  it('changes the ranking: one choice lifts a weaker match to the top', () => {
    const entries = record([], 'last', 'answer:system.load', T0);
    const load = boosted(1.5, boost(entries, 'last', 'answer:system.load', T0 + DAY));
    const lastfm = boosted(2, boost(entries, 'last', 'bookmark:lastfm', T0 + DAY));
    expect(load).toBeGreaterThan(lastfm);
    // Half a year later the choice has faded and the better match wins again.
    const later = T0 + 180 * DAY;
    expect(boosted(1.5, boost(entries, 'last', 'answer:system.load', later))).toBeLessThan(lastfm);
  });

  it('evicts the least valuable entries beyond the limit', () => {
    let entries: FrecencyEntry[] = [];
    for (let i = 0; i < MAX_ENTRIES + 20; i++) entries = record(entries, `q${i}`, `t${i}`, T0 + i * 1000);
    expect(entries).toHaveLength(MAX_ENTRIES);
    // The oldest choices (lowest decayed value) went first.
    expect(entries.some((e) => e.key === 'q0')).toBe(false);
    expect(entries.some((e) => e.key === `q${MAX_ENTRIES + 19}`)).toBe(true);
  });

  it('forgets faded entries', () => {
    const entries = record([], 'old', 't', T0);
    const later = record(entries, 'new', 'u', T0 + 10 * HALF_LIFE_MS);
    expect(later.map((e) => e.key)).toEqual(['new']);
  });

  it('ignores empty queries and targets', () => {
    expect(record([], '  ', 't', T0)).toEqual([]);
    expect(record([], 'q', '', T0)).toEqual([]);
  });

  it('validates stored data', () => {
    expect(parseEntries('nope')).toEqual([]);
    expect(
      parseEntries([
        { key: 'a', target: 't', value: 1, at: T0 },
        { key: 'b', target: 't', value: 'x', at: T0 },
        { key: 'c', target: 't', value: Infinity, at: T0 },
        null,
      ]),
    ).toEqual([{ key: 'a', target: 't', value: 1, at: T0 }]);
  });
});
