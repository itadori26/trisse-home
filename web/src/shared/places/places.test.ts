import { afterEach, describe, expect, it, vi } from 'vitest';
import { browserLocation, placeAt, placeIdOf, roundCoordinate, type SavedPlace } from './places';

const home: SavedPlace = { id: 'p1', name: 'Home', label: '', lat: 53.1435, lon: 8.2146 };

describe('saved places', () => {
  it('reads the place id from widget settings', () => {
    expect(placeIdOf({ placeId: 'p1' })).toBe('p1');
    expect(placeIdOf({ placeId: 3 })).toBe('');
    expect(placeIdOf({})).toBe('');
  });

  it('finds a place at nearly the same coordinates', () => {
    expect(placeAt([home], 53.1438, 8.2149)).toBe(home);
    expect(placeAt([home], 53.2, 8.2146)).toBeUndefined();
  });
});

describe('browser location (#693)', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('rounds coordinates to about 100 m', () => {
    expect(roundCoordinate(53.143549)).toBe(53.144);
    expect(roundCoordinate(-8.21449)).toBe(-8.214);
  });

  it('returns the rounded position', async () => {
    const getCurrentPosition = (ok: PositionCallback) =>
      ok({ coords: { latitude: 53.143549, longitude: 8.214649 } } as GeolocationPosition);
    vi.stubGlobal('navigator', { geolocation: { getCurrentPosition } });
    await expect(browserLocation()).resolves.toEqual({ lat: 53.144, lon: 8.215 });
  });

  it('rejects without geolocation', async () => {
    vi.stubGlobal('navigator', {});
    await expect(browserLocation()).rejects.toThrow();
  });
});
