import { describe, expect, it } from 'vitest';
import { callParty, formatRate } from './fritzbox';

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
});
