import { describe, expect, it } from 'vitest';
import { age, itemMeta, parseRssConfig } from './feed';

describe('rss domain', () => {
  const now = new Date('2026-11-10T12:00:00Z');
  it.each([
    ['2026-11-10T11:55:00Z', '5 min'],
    ['2026-11-10T09:00:00Z', '3 h'],
    ['2026-11-07T12:00:00Z', '3 d'],
  ])('age(%s) = %s', (d, want) => expect(age(new Date(d), now)).toBe(want));
  it('clamps the limit and defaults to unread only', () => {
    expect(parseRssConfig({ limit: 1000 })).toEqual({
      feedId: '',
      unreadOnly: true,
      limit: 100,
      compact: false,
    });
  });

  it('reads the compact option and defaults to off (#125)', () => {
    expect(parseRssConfig({}).compact).toBe(false);
    expect(parseRssConfig({ compact: true }).compact).toBe(true);
    expect(parseRssConfig({ compact: 'yes' }).compact).toBe(false);
  });

  it('builds the meta line with the feed name only for several feeds (#125)', () => {
    const item = { feedTitle: 'Go Blog', publishedAt: new Date('2026-11-10T09:00:00Z') };
    expect(itemMeta(item, false, now)).toBe('Go Blog · 3 h');
    expect(itemMeta(item, true, now)).toBe('3 h');
    expect(itemMeta({ ...item, feedTitle: '' }, false, now)).toBe('3 h');
  });
});
