import { describe, expect, it } from 'vitest';
import { minutesUntil, modeIcon, parseTransitConfig } from './transit';

describe('transit domain', () => {
  it('parses the widget config', () => {
    expect(parseTransitConfig({})).toBeNull();
    expect(parseTransitConfig({ stopId: 's', count: 99 })?.count).toBe(20);
  });
  it('counts minutes using real-time data', () => {
    const now = new Date('2026-11-11T20:00:00Z');
    const planned = new Date('2026-11-11T20:03:00Z');
    expect(minutesUntil({ planned, expected: null }, now)).toBe('3 min');
    expect(minutesUntil({ planned, expected: new Date('2026-11-11T20:10:00Z') }, now)).toBe('10 min');
    expect(minutesUntil({ planned: now, expected: null }, now)).toBe('now');
  });
  it('picks mode icons', () => {
    expect(modeIcon('bus')).toBe('🚌');
    expect(modeIcon('regional_rail')).toBe('🚆');
  });
});
