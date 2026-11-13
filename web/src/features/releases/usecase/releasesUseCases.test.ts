import { describe, expect, it, vi } from 'vitest';
import { ReleasesUseCases, type ReleasesGateway } from './releasesUseCases';

describe('ReleasesUseCases', () => {
  it('does not search for single letters', async () => {
    const g = {
      search: vi.fn().mockResolvedValue([]),
      followed: vi.fn(),
      follow: vi.fn(),
      unfollow: vi.fn(),
      list: vi.fn(),
    } satisfies ReleasesGateway;
    const uc = new ReleasesUseCases(g);
    expect(await uc.search(' a ')).toEqual([]);
    expect(g.search).not.toHaveBeenCalled();
    await uc.search(' björk ');
    expect(g.search).toHaveBeenCalledWith('björk');
  });
});
