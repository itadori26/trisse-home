import { describe, expect, it } from 'vitest';
import { faviconUrl, iconSrc, linkProblem, parseBookmarksConfig, parseTags } from './bookmark';

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
    expect(parseBookmarksConfig({ folderId: 'f', tag: 1 })).toEqual({ folderId: 'f', tag: '' });
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
