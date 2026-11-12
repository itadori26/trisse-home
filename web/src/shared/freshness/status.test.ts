import { describe, expect, it } from 'vitest';
import { combine, freshness, minutesAgo, type WidgetStatus } from './status';

const s = (fetchedAt: number | null, lastError: string | null = null, intervalMs = 60_000): WidgetStatus => ({
  fetchedAt,
  lastError,
  intervalMs,
});

describe('freshness (#94)', () => {
  it('classifies statuses', () => {
    expect(freshness(s(null), 0)).toBe('loading');
    expect(freshness(s(0), 120_000)).toBe('fresh');
    expect(freshness(s(0), 120_001)).toBe('stale');
    expect(freshness(s(0, 'down'), 1000)).toBe('error');
  });

  it('combines jobs: worst state, oldest data, first error', () => {
    const c = combine([s(100_000), s(50_000, 'timeout'), s(90_000)], 120_000);
    expect(c).toEqual({ freshness: 'error', fetchedAt: 50_000, lastError: 'timeout' });
    expect(combine([], 0).freshness).toBe('fresh');
  });

  it('counts minutes', () => {
    expect(minutesAgo(0, 59_000)).toBe(0);
    expect(minutesAgo(0, 61_000)).toBe(1);
    expect(minutesAgo(10, 0)).toBe(0);
  });
});
