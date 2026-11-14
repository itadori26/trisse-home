import { describe, expect, it, vi } from 'vitest';
import { formatBytes, looksLikeUrl, parseReadLaterConfig } from './readlater';
import { ReadLaterUseCases, type ReadLaterGateway } from '../usecase/readLaterUseCases';

describe('read later (#99)', () => {
  it('checks links, config and sizes', () => {
    expect(looksLikeUrl('https://example.org/post')).toBe(true);
    expect(looksLikeUrl('example.org')).toBe(false);
    expect(looksLikeUrl('javascript:alert(1)')).toBe(false);
    expect(parseReadLaterConfig({ count: 99 })).toEqual({ count: 30 });
    expect(parseReadLaterConfig({})).toEqual({ count: 6 });
    expect(formatBytes(512, 'en')).toBe('512 B');
    expect(formatBytes(2048, 'en')).toBe('2 KB');
    expect(formatBytes(3.5 * 1024 * 1024, 'de')).toBe('3,5 MB');
  });

  it('marks an unread article as read when it is opened', async () => {
    const article = { id: 'a', status: 'unread' } as never;
    const g = {
      list: vi.fn(),
      save: vi.fn(),
      get: vi.fn().mockResolvedValue(article),
      setStatus: vi.fn().mockResolvedValue(undefined),
      remove: vi.fn(),
      archiveDays: vi.fn().mockRejectedValue(new Error('404')),
      saveArchiveDays: vi.fn(),
      importPocket: vi.fn().mockResolvedValue({ added: 1, duplicates: 0, invalid: 0 }),
    } satisfies ReadLaterGateway;
    const uc = new ReadLaterUseCases(g);
    expect((await uc.open('a')).status).toBe('read');
    expect(g.setStatus).toHaveBeenCalledWith('a', 'read');
    await expect(uc.save('no link')).rejects.toThrow();
    expect(g.save).not.toHaveBeenCalled();
    // Not available when the backend refuses (module off or a guest), asked once.
    expect(await uc.available()).toBe(false);
    await uc.available();
    expect(g.archiveDays).toHaveBeenCalledTimes(1);
    await expect(uc.importPocket({ size: (5 << 20) + 1, text: () => Promise.resolve('') })).rejects.toThrow();
    await uc.importPocket({ size: 3, text: () => Promise.resolve('csv') });
    expect(g.importPocket).toHaveBeenCalledWith('csv');
  });
});
