import { describe, expect, it } from 'vitest';
import { formatDuration, formatUptime, uptimeLevel } from './uptime';

describe('uptime', () => {
  it('formats percentages', () => {
    expect(formatUptime(null)).toBe('–');
    expect(formatUptime(100)).toBe('100 %');
    expect(formatUptime(99.956)).toBe('99.96 %');
    expect(formatUptime(87.25)).toBe('87.3 %');
  });

  it('classifies days', () => {
    expect(uptimeLevel(null)).toBe('none');
    expect(uptimeLevel(99.9)).toBe('good');
    expect(uptimeLevel(97)).toBe('degraded');
    expect(uptimeLevel(50)).toBe('bad');
  });

  it('formats durations', () => {
    expect(formatDuration(30)).toBe('30 s');
    expect(formatDuration(45 * 60)).toBe('45 min');
    expect(formatDuration(200 * 60)).toBe('3 h 20 min');
    expect(formatDuration(3 * 86400)).toBe('3 d');
  });
});
