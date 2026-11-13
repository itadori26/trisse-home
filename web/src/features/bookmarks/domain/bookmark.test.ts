import { describe, expect, it } from 'vitest';
import { faviconUrl, iconSrc, linkProblem, parseBookmarksConfig, parseTags, sortBookmarks } from './bookmark';

describe('bookmarks domain', () => {
  it('parses tags', () => {
    expect(parseTags(' Go, web  #server go ')).toEqual(['go', 'web', 'server']);
    expect(parseTags('')).toEqual([]);
  });
  it('builds favicon urls', () => {
    expect(faviconUrl('https://go.dev/doc')).toBe('https://go.dev/favicon.ico');
    expect(faviconUrl('not a url')).toBeNull();
  });
  it('parses widget config', () => {
    expect(parseBookmarksConfig({ folderId: 'f', tag: 1 })).toEqual({
      folderId: 'f',
      tag: '',
      sort: 'manual',
    });
  });
});

describe('icons and link status (#72)', () => {
  it('prefers the stored icon', () => {
    expect(iconSrc({ id: 'b 1', url: 'https://go.dev/x', hasIcon: true })).toBe('/api/bookmarks/b%201/icon');
    expect(iconSrc({ id: 'b1', url: 'https://go.dev/x', hasIcon: false })).toBe('https://go.dev/favicon.ico');
  });
  it('describes problems', () => {
    const checkedAt = '2026-11-12T10:00:00Z';
    expect(linkProblem({ state: 'broken', code: 404, target: '', checkedAt }, 'en-US')).toContain('404');
    expect(linkProblem({ state: 'broken', code: 0, target: '', checkedAt }, 'en-US')).toMatch(
      /not reachable/,
    );
    expect(
      linkProblem({ state: 'redirect', code: 200, target: 'https://new.example', checkedAt }, 'en-US'),
    ).toContain('https://new.example');
    expect(linkProblem({ state: 'ok', code: 200, target: '', checkedAt })).toBe('');
  });
});

describe('bookmark sort order (#120)', () => {
  const list = [
    { title: 'banana', createdAt: '2026-01-02T00:00:00Z' },
    { title: 'Apple', createdAt: '2026-03-01T00:00:00Z' },
    { title: 'server 10', createdAt: '2026-02-01T00:00:00Z' },
    { title: 'Server 2', createdAt: '2026-02-01T00:00:00Z' },
  ];
  const titles = (l: { title: string }[]) => l.map((b) => b.title);

  it('defaults to the server order and rejects unknown values', () => {
    expect(parseBookmarksConfig({}).sort).toBe('manual');
    expect(parseBookmarksConfig({ sort: 'az' }).sort).toBe('az');
    expect(parseBookmarksConfig({ sort: 'random' }).sort).toBe('manual');
  });

  it('keeps the given order for manual and never changes the input', () => {
    const copy = structuredClone(list);
    expect(titles(sortBookmarks(list, 'manual'))).toEqual(titles(list));
    sortBookmarks(list, 'az');
    sortBookmarks(list, 'newest');
    expect(list).toEqual(copy);
  });

  it('sorts A-Z ignoring case and comparing numbers by value', () => {
    expect(titles(sortBookmarks(list, 'az', 'en'))).toEqual(['Apple', 'banana', 'Server 2', 'server 10']);
  });

  it('sorts newest first and breaks ties alphabetically', () => {
    expect(titles(sortBookmarks(list, 'newest', 'en'))).toEqual(['Apple', 'Server 2', 'server 10', 'banana']);
  });

  it('puts bookmarks with an unreadable date last', () => {
    const odd = [{ title: 'x', createdAt: '' }, ...list];
    expect(titles(sortBookmarks(odd, 'newest', 'en')).at(-1)).toBe('x');
  });
});
