import { describe, expect, it, vi } from 'vitest';
import { ProxmoxUseCases, type ProxmoxGateway } from './proxmoxUseCases';

describe('ProxmoxUseCases', () => {
  it('trims everything but the secret', async () => {
    const gateway = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      status: vi.fn(),
      act: vi.fn(),
    } satisfies ProxmoxGateway;
    await new ProxmoxUseCases(gateway).saveSettings({
      url: ' https://pve:8006 ',
      tokenId: ' a@pve!t ',
      secret: ' s ',
      fingerprint: ' AB:CD ',
    });
    expect(gateway.saveSettings).toHaveBeenCalledWith({
      url: 'https://pve:8006',
      tokenId: 'a@pve!t',
      secret: ' s ',
      fingerprint: 'AB:CD',
    });
  });
});
