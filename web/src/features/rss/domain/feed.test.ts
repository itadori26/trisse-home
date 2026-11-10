import { describe, expect, it } from 'vitest';
import { age, parseRssConfig } from './feed';

describe('rss domain', () => {
  const now = new Date('2026-11-10T12:00:00Z');
  it.each([
    ['2026-11-10T11:55:00Z', '5 min'],
    ['2026-11-10T09:00:00Z', '3 h'],
    ['2026-11-07T12:00:00Z', '3 d'],
  ])('age(%s) = %s', (d, want) => expect(age(new Date(d), now)).toBe(want));
  it('clamps the limit and defaults to unread only', () => {
    expect(parseRssConfig({ limit: 1000 })).toEqual({ feedId: '', unreadOnly: true, limit: 100 });
  });
});
