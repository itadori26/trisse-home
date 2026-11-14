import { describe, expect, it } from 'vitest';
import { callParty, callsOptions, formatRate, internetOptions } from './fritzbox';

describe('fritzbox domain', () => {
  it('formats rates', () => {
    expect(formatRate(12_500_000)).toBe('100 Mbit/s');
    expect(formatRate(500_000)).toBe('4.0 Mbit/s');
    expect(formatRate(1000)).toBe('8 kbit/s');
  });
  it('names the other party', () => {
    expect(callParty({ name: 'Oma', number: '0441' })).toBe('Oma');
    expect(callParty({ name: '', number: '0441' })).toBe('0441');
    expect(callParty({ name: '', number: '' })).toBe('Unknown');
  });
  it('hides numbers when asked (#168)', () => {
    expect(callParty({ name: 'Oma', number: '0441' }, false)).toBe('Oma');
    expect(callParty({ name: '', number: '0441' }, false)).toBe('Unknown');
  });
  it('reads the widget options with defaults (#168)', () => {
    expect(internetOptions({})).toEqual({ showIp: true, showRates: true, showUptime: true });
    expect(internetOptions({ showIp: false })).toMatchObject({ showIp: false, showRates: true });
    expect(callsOptions({})).toEqual({ filter: 'all', limit: 10, showNumbers: true });
    expect(callsOptions({ filter: 'missed', limit: 25, showNumbers: false })).toEqual({
      filter: 'missed',
      limit: 25,
      showNumbers: false,
    });
    expect(callsOptions({ filter: 'rejected', limit: 3 })).toMatchObject({ filter: 'all', limit: 10 });
    expect(callsOptions({ limit: 51 }).limit).toBe(10);
  });
});
