import { describe, expect, it } from 'vitest';
import { overallStatus, powerOnAge, type Disk, type DiskStatus } from './disk';

const disk = (status: DiskStatus): Disk => ({
  device: '/dev/sda',
  model: '',
  protocol: 'ATA',
  capacityBytes: 0,
  checkedAt: new Date(0),
  status,
  findings: [],
  temperature: null,
  powerOnHours: null,
  reallocated: null,
  pending: null,
  percentageUsed: null,
});

describe('disk health', () => {
  it('takes the worst status', () => {
    expect(overallStatus([])).toBe('ok');
    expect(overallStatus([disk('ok'), disk('unknown')])).toBe('unknown');
    expect(overallStatus([disk('critical'), disk('warning')])).toBe('critical');
  });

  it('formats power-on hours as years and days', () => {
    expect(powerOnAge(33712)).toEqual({ years: 3, days: 309 });
    expect(powerOnAge(10)).toEqual({ years: 0, days: 0 });
  });
});
