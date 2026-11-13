import { describe, expect, it } from 'vitest';
import { availableAction, parseWidgetConfig, usage, visibleGuests, type ProxmoxGuest } from './proxmox';

const guest = (id: number, status: string, actions = false): ProxmoxGuest => ({
  id,
  name: `g${id}`,
  node: 'pve',
  kind: 'qemu',
  status,
  uptimeSeconds: 0,
  cpuPercent: 0,
  memUsed: 0,
  memTotal: 0,
  actions,
});

describe('proxmox domain', () => {
  it('computes usage without dividing by zero', () => {
    expect(usage(1, 4)).toBe(25);
    expect(usage(1, 0)).toBe(0);
  });
  it('offers start for stopped and shutdown for running allow-listed guests', () => {
    expect(availableAction(guest(1, 'running', true))).toBe('shutdown');
    expect(availableAction(guest(1, 'stopped', true))).toBe('start');
    expect(availableAction(guest(1, 'paused', true))).toBeNull();
    expect(availableAction(guest(1, 'running'))).toBeNull();
  });
  it('filters running guests when configured', () => {
    const list = [guest(1, 'running'), guest(2, 'stopped')];
    expect(visibleGuests(list, parseWidgetConfig({}))).toHaveLength(2);
    expect(visibleGuests(list, parseWidgetConfig({ runningOnly: true })).map((g) => g.id)).toEqual([1]);
  });
});
