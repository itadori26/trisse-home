import { describe, expect, it, vi } from 'vitest';
import { AdGuardUseCases, type AdGuardGateway } from './adguardUseCases';

describe('AdGuardUseCases', () => {
  it('trims url and username but not the password', async () => {
    const gateway = { settings: vi.fn(), saveSettings: vi.fn(), stats: vi.fn() } satisfies AdGuardGateway;
    await new AdGuardUseCases(gateway).saveSettings(' http://ag:3000 ', ' admin ', ' pw ');
    expect(gateway.saveSettings).toHaveBeenCalledWith('http://ag:3000', 'admin', ' pw ');
  });
});
