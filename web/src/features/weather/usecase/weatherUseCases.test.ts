import { describe, expect, it, vi } from 'vitest';
import { type Forecast, noHealth } from '../domain/weather';
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
  windDirection: 315,
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
    health: vi.fn(),
    healthOptions: vi.fn(),
  };
  const uc = new WeatherUseCases(gateway);

  it('limits the number of days', async () => {
    expect(
      (await uc.forecast({ lat: 1, lon: 2, label: '', days: 2, warnings: true, ...noHealth })).days,
    ).toHaveLength(2);
  });
  it('shares a forecast for a minute between callers of the same place', async () => {
    let clock = 0;
    const gw: WeatherGateway = { ...gateway, forecast: vi.fn().mockResolvedValue(forecast) };
    const shared = new WeatherUseCases(gw, () => clock);
    const cfg = { lat: 1, lon: 2, label: '', days: 2, warnings: false, ...noHealth };
    await Promise.all([shared.forecast(cfg), shared.forecast({ ...cfg, days: 5 })]);
    expect(gw.forecast).toHaveBeenCalledTimes(1);
    await shared.forecast({ ...cfg, lat: 3 });
    expect(gw.forecast).toHaveBeenCalledTimes(2);
    clock = 61_000;
    await shared.forecast(cfg);
    expect(gw.forecast).toHaveBeenCalledTimes(3);
  });
  it('does not share a failed request', async () => {
    const gw: WeatherGateway = {
      ...gateway,
      forecast: vi.fn().mockRejectedValueOnce(new Error('down')).mockResolvedValue(forecast),
    };
    const shared = new WeatherUseCases(gw, () => 0);
    const cfg = { lat: 1, lon: 2, label: '', days: 2, warnings: false, ...noHealth };
    await expect(shared.forecast(cfg)).rejects.toThrow('down');
    await expect(shared.forecast(cfg)).resolves.toBeDefined();
  });
  it('skips warnings when disabled', async () => {
    await uc.warnings({ lat: 1, lon: 2, label: '', days: 2, warnings: false, ...noHealth });
    expect(gateway.warnings).not.toHaveBeenCalled();
    await uc.warnings({ lat: 1, lon: 2, label: '', days: 2, warnings: true, ...noHealth });
    expect(gateway.warnings).toHaveBeenCalledWith(1, 2);
  });
  it('does not search for single characters', async () => {
    await uc.searchPlaces(' a ');
    expect(gateway.searchPlaces).not.toHaveBeenCalled();
  });

  it('asks for pollen and UV only when configured (#56)', async () => {
    const g = {
      forecast: vi.fn(),
      searchPlaces: vi.fn(),
      warnings: vi.fn(),
      radarUrl: vi.fn(),
      health: vi.fn().mockResolvedValue({ pollen: null, uv: null }),
      healthOptions: vi.fn(),
    } satisfies WeatherGateway;
    const uc = new WeatherUseCases(g);
    const cfg = { lat: 1, lon: 2, label: '', days: 2, warnings: false, ...noHealth };
    expect(await uc.health(cfg)).toBeNull();
    expect(g.health).not.toHaveBeenCalled();
    await uc.health({ ...cfg, pollenRegion: 32, pollenTypes: ['birch'] });
    expect(g.health).toHaveBeenCalledWith(32, '', ['birch']);
    await uc.health({ ...cfg, uvCity: 'Hannover' });
    expect(g.health).toHaveBeenLastCalledWith(0, 'Hannover', []);
  });
});
