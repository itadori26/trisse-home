import { describe, expect, it, vi } from 'vitest';
import { NextcloudUseCases, type NextcloudGateway } from './nextcloudUseCases';

describe('NextcloudUseCases', () => {
  it('trims the settings before saving', async () => {
    const gateway = { settings: vi.fn(), saveSettings: vi.fn(), status: vi.fn() } satisfies NextcloudGateway;
    await new NextcloudUseCases(gateway).saveSettings(' https://c ', ' anna ', ' pw ');
    expect(gateway.saveSettings).toHaveBeenCalledWith('https://c', 'anna', 'pw');
  });
});
