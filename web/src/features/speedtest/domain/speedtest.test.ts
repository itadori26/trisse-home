import { describe, expect, it } from 'vitest';
import { belowThreshold, formatMbps } from './speedtest';

describe('speedtest', () => {
  it('formats speeds', () => {
    expect(formatMbps(250.4)).toBe('250');
    expect(formatMbps(42.35)).toBe('42.4');
  });

  it('checks thresholds', () => {
    const m = { at: new Date(), downloadMbps: 80, uploadMbps: 30, pingMs: 10 };
    expect(belowThreshold(m, { minDownMbps: 100, minUpMbps: 0 })).toBe(true);
    expect(belowThreshold(m, { minDownMbps: 0, minUpMbps: 0 })).toBe(false);
    expect(belowThreshold(null, { minDownMbps: 100, minUpMbps: 0 })).toBe(false);
  });
});
