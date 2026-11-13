import { describe, expect, it, vi } from 'vitest';
import { MediaUseCases, type MediaGateway } from './mediaUseCases';

describe('MediaUseCases', () => {
  it('trims url and token', async () => {
    const g = { settings: vi.fn(), saveSettings: vi.fn(), overview: vi.fn() } satisfies MediaGateway;
    await new MediaUseCases(g).saveSettings('plex', ' http://plex:32400 ', ' tok ');
    expect(g.saveSettings).toHaveBeenCalledWith('plex', 'http://plex:32400', 'tok');
  });
});
