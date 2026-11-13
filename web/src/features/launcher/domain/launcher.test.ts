import { describe, expect, it } from 'vitest';
import {
  buildView,
  initial,
  type LauncherBookmark,
  type LauncherService,
  matches,
  safeUrl,
  SERVICE_PREVIEW,
  statusSummary,
} from './launcher';

const services: LauncherService[] = Array.from({ length: 8 }, (_, i) => ({
  id: `s${i}`,
  name: i === 0 ? 'Nextcloud' : `Service ${i}`,
  url: `https://s${i}.lan`,
  state: i === 3 ? 'offline' : 'online',
  latencyMs: 10 + i,
}));
const bookmarks: LauncherBookmark[] = [
  { id: 'b1', title: 'GitHub', url: 'https://github.com', folder: 'Dev' },
  { id: 'b2', title: 'Wikipedia', url: 'https://wikipedia.org', folder: 'Wissen' },
];
const opts = {
  filter: '',
  segment: 'all' as const,
  content: 'both' as const,
  expanded: new Set<'services' | 'bookmarks'>(),
};

describe('launcher', () => {
  it('shows the first rows with "show all" only when entries are hidden', () => {
    const v = buildView(services, bookmarks, opts);
    expect(v.services?.items).toHaveLength(SERVICE_PREVIEW);
    expect(v.services?.more).toBe(true);
    expect(v.bookmarks?.more).toBe(false);
    expect(v.counts).toEqual({ all: 10, services: 8, bookmarks: 2 });
    const all = buildView(services, bookmarks, { ...opts, expanded: new Set(['services']) });
    expect(all.services?.items).toHaveLength(8);
    expect(all.services?.more).toBe(false);
  });

  it('filters across both kinds, also by folder and without accents', () => {
    expect(buildView(services, bookmarks, { ...opts, filter: 'next' }).counts).toEqual({
      all: 1,
      services: 1,
      bookmarks: 0,
    });
    expect(buildView(services, bookmarks, { ...opts, filter: 'wissen' }).bookmarks?.items[0]?.title).toBe(
      'Wikipedia',
    );
    expect(matches('wiß', 'Wiß')).toBe(true);
    expect(matches('git hub', 'GitHub')).toBe(true);
    expect(matches('gitlab', 'GitHub')).toBe(false);
  });

  it('segments and the content setting', () => {
    const seg = buildView(services, bookmarks, { ...opts, segment: 'services' });
    expect(seg.bookmarks).toBeNull();
    expect(seg.services?.items).toHaveLength(8); // a segment alone shows everything
    const onlyBookmarks = buildView(services, bookmarks, { ...opts, content: 'bookmarks' });
    expect(onlyBookmarks.services).toBeNull();
    expect(onlyBookmarks.counts.all).toBe(2);
  });

  it('summarises the service status', () => {
    expect(statusSummary([])).toEqual({ kind: 'none', offline: 0 });
    expect(statusSummary(services)).toEqual({ kind: 'offline', offline: 1 });
    expect(statusSummary([{ state: 'online' }, { state: 'unknown' }])).toEqual({ kind: 'ok', offline: 0 });
  });

  it('only opens http(s) links and builds initials', () => {
    expect(safeUrl('https://a.example/x')).toBe('https://a.example/x');
    expect(safeUrl('javascript:alert(1)')).toBeNull();
    expect(safeUrl('not a url')).toBeNull();
    expect(initial(' über')).toBe('Ü');
    expect(initial('42 news')).toBe('4');
    expect(initial('—')).toBe('?');
  });
});
