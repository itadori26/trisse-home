import { describe, expect, it } from 'vitest';
import { entryName, parseHolidaysConfig, periodDays, stateName } from './holidays';

describe('holidays domain (#55)', () => {
  it('reads the widget config with defaults', () => {
    expect(parseHolidaysConfig({})).toEqual({ state: 'NI', count: 6 });
    expect(parseHolidaysConfig({ state: 'BY', count: 50 })).toEqual({ state: 'BY', count: 20 });
    expect(parseHolidaysConfig({ state: 'XX', count: 0 })).toEqual({ state: 'NI', count: 1 });
  });

  it('names entries and states', () => {
    expect(entryName({ kind: 'public', key: 'good_friday', name: '' })).toBe('Good Friday');
    expect(entryName({ kind: 'public', key: 'unknown_day', name: '' })).toBe('unknown_day');
    expect(entryName({ kind: 'school', key: '', name: 'Sommerferien' })).toBe('Sommerferien');
    expect(stateName('NI')).toBe('Lower Saxony');
  });

  it('counts the days of a period inclusively', () => {
    expect(periodDays({ start: new Date(2026, 11, 23), end: new Date(2027, 0, 5) })).toBe(14);
    expect(periodDays({ start: new Date(2027, 4, 7), end: new Date(2027, 4, 7) })).toBe(1);
  });
});
