import { describe as group, expect, it } from 'vitest';
import { describe, parseWeatherConfig, placeLabel } from './weather';

group('weather domain', () => {
  it('parses a valid config and clamps days', () => {
    expect(parseWeatherConfig({ lat: 53.1, lon: 8.2, label: 'Home', days: 9 })).toEqual({
      lat: 53.1,
      lon: 8.2,
      label: 'Home',
      days: 5,
    });
  });
  it('returns null without coordinates', () => {
    expect(parseWeatherConfig({})).toBeNull();
    expect(parseWeatherConfig({ lat: '53', lon: 8 })).toBeNull();
  });
  it('describes WMO codes including night', () => {
    expect(describe(63).label).toBe('Rain');
    expect(describe(0, false).icon).toBe('🌙');
    expect(describe(1234).label).toBe('Unknown');
  });
  it('builds place labels without duplicates', () => {
    expect(placeLabel({ name: 'Berlin', region: 'Berlin', country: 'Germany', lat: 0, lon: 0 })).toBe(
      'Berlin, Germany',
    );
  });
});
