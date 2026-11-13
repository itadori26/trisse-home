import { describe, expect, it, vi } from 'vitest';
import type { BriefingSettings } from '../domain/briefing';
import { type BriefingGateway, BriefingUseCases, BriefingValidationError } from './briefingUseCases';

describe('BriefingUseCases', () => {
  const gateway = {
    get: vi.fn(),
    settings: vi.fn(),
    saveSettings: vi.fn().mockImplementation((s: BriefingSettings) => Promise.resolve(s)),
    sendNow: vi.fn(),
  } satisfies BriefingGateway;
  const uc = new BriefingUseCases(gateway);
  const s: BriefingSettings = {
    layout: [{ kind: 'todos', enabled: true }],
    pushOn: true,
    pushAt: '06:30',
    pushDays: [5, 1, 5],
    skipEmpty: true,
    lat: 53.1,
    lon: 8.8,
    available: ['todos'],
  };

  it('saves valid settings with de-duplicated weekdays', async () => {
    await uc.saveSettings(s);
    expect(gateway.saveSettings).toHaveBeenCalledWith({ ...s, pushDays: [1, 5] });
  });

  it('rejects invalid settings without calling the backend', async () => {
    gateway.saveSettings.mockClear();
    await expect(uc.saveSettings({ ...s, pushAt: '' })).rejects.toBeInstanceOf(BriefingValidationError);
    expect(gateway.saveSettings).not.toHaveBeenCalled();
  });
});
