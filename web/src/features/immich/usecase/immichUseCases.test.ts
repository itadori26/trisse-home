import { describe, expect, it, vi } from 'vitest';
import { ImmichUseCases, type ImmichGateway } from './immichUseCases';

describe('ImmichUseCases', () => {
  it('trims url and key', async () => {
    const g = { settings: vi.fn(), saveSettings: vi.fn(), overview: vi.fn() } satisfies ImmichGateway;
    await new ImmichUseCases(g).saveSettings(' http://immich:2283 ', ' k ');
    expect(g.saveSettings).toHaveBeenCalledWith('http://immich:2283', 'k');
  });
});
