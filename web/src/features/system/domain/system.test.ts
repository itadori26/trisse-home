import { describe, expect, it } from 'vitest';
import { daysUntilFull, loadPercent, parseInstallMethod, updateHintKey, type SystemSnapshot } from './system';

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

describe('install method', () => {
  it('accepts the known methods', () => {
    expect(parseInstallMethod('docker')).toBe('docker');
    expect(parseInstallMethod('package')).toBe('package');
    expect(parseInstallMethod('appliance')).toBe('appliance');
  });
  it('maps anything else to unknown', () => {
    expect(parseInstallMethod('')).toBe('unknown');
    expect(parseInstallMethod('snap')).toBe('unknown');
  });
  it('has an update hint per method', () => {
    expect(updateHintKey('docker')).toBe('system.update_how_docker');
    expect(updateHintKey('package')).toBe('system.update_how_package');
    expect(updateHintKey('appliance')).toBe('system.update_how_appliance');
    expect(updateHintKey('unknown')).toBeNull();
  });
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
