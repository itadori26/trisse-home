import { describe, expect, it, vi } from 'vitest';
import { ParcelsUseCases, type ParcelsGateway } from './parcelsUseCases';

describe('ParcelsUseCases', () => {
  it('validates before adding', async () => {
    const gateway = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      list: vi.fn(),
      add: vi.fn(),
      remove: vi.fn(),
      mailSettings: vi.fn(),
      saveMailSettings: vi.fn(),
    } satisfies ParcelsGateway;
    const uc = new ParcelsUseCases(gateway);
    await expect(uc.add('x', '')).rejects.toThrow(/8 to 35/);
    await uc.add('0034 0434 1234', ' Shoes ');
    expect(gateway.add).toHaveBeenCalledWith('003404341234', 'dhl', 'Shoes');
    await uc.add('12345678901', '', 'gls');
    expect(gateway.add).toHaveBeenLastCalledWith('12345678901', 'gls', '');
    await uc.saveMailSettings({
      enabled: true,
      host: ' imap.example.com ',
      port: 993,
      username: ' me ',
      password: ' pw ',
      folder: ' Pakete ',
      intervalMinutes: 15,
    });
    expect(gateway.saveMailSettings).toHaveBeenCalledWith({
      enabled: true,
      host: 'imap.example.com',
      port: 993,
      username: 'me',
      password: ' pw ',
      folder: 'Pakete',
      intervalMinutes: 15,
    });
  });
});
