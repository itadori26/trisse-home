import { describe, expect, it, vi } from 'vitest';
import { GamedealsUseCases, type GameDealsGateway } from './gamedealsUseCases';

describe('GamedealsUseCases', () => {
  it('normalises the Steam ID and country', async () => {
    const g = { settings: vi.fn(), saveSettings: vi.fn(), overview: vi.fn() } satisfies GameDealsGateway;
    await new GamedealsUseCases(g).saveSettings({
      steamId: 'https://steamcommunity.com/profiles/76561197960287930',
      country: ' de ',
      minDiscount: 20,
      epic: true,
    });
    expect(g.saveSettings).toHaveBeenCalledWith({
      steamId: '76561197960287930',
      country: 'DE',
      minDiscount: 20,
      epic: true,
    });
  });
});
