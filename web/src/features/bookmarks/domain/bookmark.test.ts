import { describe, expect, it } from 'vitest';
import { faviconUrl, parseBookmarksConfig, parseTags } from './bookmark';

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
