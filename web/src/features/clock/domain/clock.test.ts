import { describe, expect, it } from 'vitest';
import { formatTime, isValidTimeZone, parseClockConfig } from './clock';

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
});
