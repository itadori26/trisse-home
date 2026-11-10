import { describe, expect, it } from 'vitest';
import { loadPercent, type SystemSnapshot } from './system';

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
