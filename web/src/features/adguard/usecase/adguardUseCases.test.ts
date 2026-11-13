import { describe, expect, it, vi } from 'vitest';
import { AdGuardUseCases, type AdGuardGateway } from './adguardUseCases';

const gateway = () => ({ settings: vi.fn(), saveSettings: vi.fn(), stats: vi.fn() }) satisfies AdGuardGateway;

describe('AdGuardUseCases', () => {
  it('trims url and username but not the password', async () => {
    const g = gateway();
    await new AdGuardUseCases(g).saveSettings('adguard', ' http://ag:3000 ', ' admin ', ' pw ');
    expect(g.saveSettings).toHaveBeenCalledWith('adguard', 'http://ag:3000', 'admin', ' pw ');
  });
  it('defaults to AdGuard Home and passes the provider through', async () => {
    const g = gateway();
    const uc = new AdGuardUseCases(g);
    await uc.stats();
    await uc.stats('pihole');
    await uc.settings('pihole');
    expect(g.stats.mock.calls).toEqual([['adguard'], ['pihole']]);
    expect(g.settings).toHaveBeenCalledWith('pihole');
  });
});
