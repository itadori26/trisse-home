import { describe, expect, it } from 'vitest';
import { daysUntilFull, loadPercent, type SystemSnapshot } from './system';

const base: SystemSnapshot = {
  cpuPercent: 0,
  cpuCount: 4,
  load: [2, 1, 1],
  memory: { total: 0, used: 0, percent: 0 },
  swap: { total: 0, used: 0, percent: 0 },
  disks: [],
  uptimeSeconds: 0,
  temperature: 0,
};

describe('loadPercent', () => {
  it('normalises by cpu count', () => expect(loadPercent(base)).toBe(50));
  it('caps at 100', () => expect(loadPercent({ ...base, load: [9, 0, 0] })).toBe(100));
  it('handles missing cpu count', () => expect(loadPercent({ ...base, cpuCount: 0 })).toBe(0));
});

describe('daysUntilFull', () => {
  const f = (status: 'filling' | 'flat' | 'insufficient', days = 0) => ({
    status,
    daysUntilFull: days,
    bytesPerDay: 0,
    confidence: 1,
  });
  it('shows only growing disks', () => {
    expect(daysUntilFull(null)).toBeNull();
    expect(daysUntilFull(f('flat'))).toBeNull();
    expect(daysUntilFull(f('insufficient'))).toBeNull();
    expect(daysUntilFull(f('filling', 22.6))).toBe(23);
  });
});
