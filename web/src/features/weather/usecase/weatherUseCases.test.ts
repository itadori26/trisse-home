import { describe, expect, it, vi } from 'vitest';
import type { Forecast } from '../domain/weather';
import { WeatherUseCases, type WeatherGateway } from './weatherUseCases';

const day = (n: number) => ({
  date: new Date(2026, 0, n),
  min: 0,
  max: 1,
  code: 0,
  precipitationProbability: 0,
});
const forecast: Forecast = {
  temperature: 1,
  apparentTemperature: 0,
  humidity: 50,
  windSpeed: 3,
  code: 0,
  isDay: true,
  days: [day(1), day(2), day(3), day(4), day(5)],
  updatedAt: new Date(),
};

describe('WeatherUseCases', () => {
  const gateway: WeatherGateway = {
    forecast: vi.fn().mockResolvedValue(forecast),
    searchPlaces: vi.fn().mockResolvedValue([]),
    warnings: vi.fn().mockResolvedValue([]),
    radarUrl: vi.fn().mockReturnValue('/radar'),
  };
  const uc = new WeatherUseCases(gateway);

  it('limits the number of days', async () => {
    expect((await uc.forecast({ lat: 1, lon: 2, label: '', days: 2, warnings: true })).days).toHaveLength(2);
  });
  it('skips warnings when disabled', async () => {
    await uc.warnings({ lat: 1, lon: 2, label: '', days: 2, warnings: false });
    expect(gateway.warnings).not.toHaveBeenCalled();
    await uc.warnings({ lat: 1, lon: 2, label: '', days: 2, warnings: true });
    expect(gateway.warnings).toHaveBeenCalledWith(1, 2);
  });
  it('does not search for single characters', async () => {
    await uc.searchPlaces(' a ');
    expect(gateway.searchPlaces).not.toHaveBeenCalled();
  });
});
