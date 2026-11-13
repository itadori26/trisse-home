import { describe, expect, it } from 'vitest';
import { LauncherUseCases } from './launcherUseCases';

const service = { id: 's', name: 'NAS', url: 'https://nas', state: 'online' as const, latencyMs: 3 };
const bookmark = { id: 'b', title: 'Docs', url: 'https://docs', folder: '' };

describe('LauncherUseCases', () => {
  it('loads both sources and survives a failing one', async () => {
    const uc = new LauncherUseCases({
      services: async () => [service],
      bookmarks: async () => {
        throw new Error('down');
      },
    });
    expect(await uc.load()).toEqual({ services: [service], bookmarks: [] });
    expect(uc.hasServices).toBe(true);
  });

  it('treats disabled modules as empty', async () => {
    const uc = new LauncherUseCases({ services: null, bookmarks: async () => [bookmark] });
    expect(await uc.load()).toEqual({ services: [], bookmarks: [bookmark] });
    expect(uc.hasServices).toBe(false);
    expect(await uc.services()).toEqual([]);
  });
});
