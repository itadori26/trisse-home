import { describe, expect, it, vi } from 'vitest';
import { FuelUseCases, type FuelGateway } from './fuelUseCases';

describe('FuelUseCases', () => {
  it('validates the location before saving', async () => {
    const g = {
      settings: vi.fn(),
      saveSettings: vi.fn(),
      setFavourites: vi.fn(),
      overview: vi.fn(),
    } satisfies FuelGateway;
    const uc = new FuelUseCases(g);
    const input = { apiKey: ' k ', lat: 0, lon: 0, place: ' x ', radiusKm: 5, fuel: 'e10' as const };
    await expect(uc.saveSettings(input)).rejects.toThrow();
    expect(g.saveSettings).not.toHaveBeenCalled();
    await uc.saveSettings({ ...input, lat: 52.5, lon: 13.4 });
    expect(g.saveSettings).toHaveBeenCalledWith({ ...input, lat: 52.5, lon: 13.4, apiKey: 'k', place: 'x' });
  });
});
