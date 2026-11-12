import { describe, expect, it } from 'vitest';
import { isUnconfigured, steamIdFrom } from './gamedeals';

describe('game deals domain', () => {
  it('extracts the Steam ID from a profile URL', () => {
    expect(steamIdFrom('https://steamcommunity.com/profiles/76561197960287930/')).toBe('76561197960287930');
    expect(steamIdFrom(' 76561197960287930 ')).toBe('76561197960287930');
  });
  it('knows when nothing is enabled', () => {
    expect(isUnconfigured({ steamId: '', epic: false })).toBe(true);
    expect(isUnconfigured({ steamId: '', epic: true })).toBe(false);
  });
});
