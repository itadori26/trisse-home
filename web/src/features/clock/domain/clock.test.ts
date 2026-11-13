import { describe, expect, it } from 'vitest';
import { formatTime, isoWeek, isValidTimeZone, parseClockConfig } from './clock';

describe('clock', () => {
  const date = new Date('2026-11-10T17:05:09Z');

  it('formats in the configured time zone', () => {
    const cfg = parseClockConfig({ timeZone: 'Europe/Berlin', showSeconds: true });
    expect(formatTime(date, cfg, 'en-GB')).toBe('18:05:09');
  });

  it('supports 12-hour format', () => {
    const cfg = parseClockConfig({ timeZone: 'UTC', hour12: true });
    expect(formatTime(date, cfg, 'en-US')).toMatch(/05:05\s?PM/);
  });

  it('falls back to defaults for invalid stored values', () => {
    const cfg = parseClockConfig({ timeZone: 'Mars/Olympus', hour12: 'yes' });
    expect(cfg.timeZone).toBe('');
    expect(cfg.hour12).toBe(false);
  });

  it('validates time zones', () => {
    expect(isValidTimeZone('Europe/Berlin')).toBe(true);
    expect(isValidTimeZone('Nope/Nope')).toBe(false);
  });

  it('shows the calendar week only when asked to', () => {
    expect(parseClockConfig({}).showWeek).toBe(false);
    expect(parseClockConfig({ showWeek: true }).showWeek).toBe(true);
    expect(parseClockConfig({ showWeek: 'yes' }).showWeek).toBe(false);
  });
});

describe('isoWeek (#122)', () => {
  const utc = { timeZone: 'UTC' };
  const week = (iso: string, tz = 'UTC') => isoWeek(new Date(iso), { timeZone: tz });

  it('counts weeks from the one with the first Thursday', () => {
    expect(week('2026-11-13T12:00:00Z')).toBe(46);
    expect(week('2026-01-01T12:00:00Z')).toBe(1);
    expect(week('2026-01-04T12:00:00Z')).toBe(1); // Sunday
    expect(week('2026-01-05T12:00:00Z')).toBe(2); // Monday
  });

  it('handles the year boundaries', () => {
    expect(week('2020-12-31T12:00:00Z')).toBe(53);
    expect(week('2021-01-01T12:00:00Z')).toBe(53);
    expect(week('2021-01-03T12:00:00Z')).toBe(53);
    expect(week('2021-01-04T12:00:00Z')).toBe(1);
    expect(week('2024-12-30T12:00:00Z')).toBe(1); // belongs to 2025
    expect(week('2015-12-31T12:00:00Z')).toBe(53);
    expect(week('2016-01-03T12:00:00Z')).toBe(53);
    expect(week('2027-01-03T12:00:00Z')).toBe(53); // 2026 has 53 weeks
    expect(isoWeek(new Date('2026-06-15T00:00:00Z'), utc)).toBe(25);
  });

  it('uses the day in the configured time zone', () => {
    // Sunday 23:30 UTC is already Monday in Berlin: the next week starts there.
    expect(week('2026-11-15T23:30:00Z', 'UTC')).toBe(46);
    expect(week('2026-11-15T23:30:00Z', 'Europe/Berlin')).toBe(47);
    expect(week('2026-11-16T01:00:00Z', 'America/Los_Angeles')).toBe(46);
  });
});
