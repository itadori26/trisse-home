import { describe, expect, it } from 'vitest';
import { normaliseTrackingNumber, parcelName } from './parcel';

describe('parcel domain', () => {
  it('normalises tracking numbers', () => {
    expect(normaliseTrackingNumber(' jjd 0003 9000 7912 345 ')).toBe('JJD000390007912345');
    expect(normaliseTrackingNumber('abc')).toBeNull();
  });
  it('names parcels', () => {
    expect(parcelName({ label: '', trackingNumber: 'X1' })).toBe('X1');
    expect(parcelName({ label: 'Shoes', trackingNumber: 'X1' })).toBe('Shoes');
  });
});
