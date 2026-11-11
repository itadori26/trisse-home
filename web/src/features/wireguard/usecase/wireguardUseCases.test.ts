import { describe, expect, it, vi } from 'vitest';
import { WireGuardUseCases, type WireGuardGateway } from './wireguardUseCases';

describe('WireGuardUseCases', () => {
  it('trims names', async () => {
    const gateway = { peers: vi.fn(), rename: vi.fn() } satisfies WireGuardGateway;
    await new WireGuardUseCases(gateway).rename('key=', '  Phone ');
    expect(gateway.rename).toHaveBeenCalledWith('key=', 'Phone');
  });
});
