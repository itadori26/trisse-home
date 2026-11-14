import { describe, expect, it, vi } from 'vitest';
import type { FeedItem } from '../domain/feed';
import { RssUseCases, type RssGateway } from './rssUseCases';

describe('RssUseCases', () => {
  const gateway = {
    feeds: vi.fn(),
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
    items: vi.fn(),
    setRead: vi.fn().mockResolvedValue(undefined),
    markAllRead: vi.fn(),
    refresh: vi.fn().mockResolvedValue([]),
    importOpml: vi.fn().mockResolvedValue({ added: 2, duplicates: 0, invalid: [] }),
    opmlUrl: vi.fn().mockReturnValue('/api/rss/opml'),
  } satisfies RssGateway;
  const uc = new RssUseCases(gateway);
  const item: FeedItem = {
    id: 'i',
    feedId: 'f',
    feedTitle: 'F',
    title: 't',
    link: '',
    publishedAt: new Date(),
    read: false,
  };

  it('marks unread items as read when opened', async () => {
    expect((await uc.open(item)).read).toBe(true);
    expect(gateway.setRead).toHaveBeenCalledWith('i', true);
  });
  it('does not call the backend for read items', async () => {
    gateway.setRead.mockClear();
    await uc.open({ ...item, read: true });
    expect(gateway.setRead).not.toHaveBeenCalled();
  });

  it('imports OPML and then loads the new feeds (#97)', async () => {
    const file = { size: 10, text: () => Promise.resolve('<opml/>') };
    const r = await uc.importOpml(file);
    expect(gateway.importOpml).toHaveBeenCalledWith('<opml/>');
    expect(gateway.refresh).toHaveBeenCalled();
    expect(r.report.added).toBe(2);
    await expect(uc.importOpml({ size: (2 << 20) + 1, text: () => Promise.resolve('') })).rejects.toThrow();
  });
});
