import { describe, expect, it } from 'vitest';
import { sparklinePoints } from './sparkline';

describe('sparklinePoints', () => {
  it('scales values into the box', () => {
    expect(sparklinePoints([0, 50, 100], 100, 10, 100)).toBe('0,10 50,5 100,0');
  });
  it('handles empty and single values', () => {
    expect(sparklinePoints([], 100, 10)).toBe('');
    expect(sparklinePoints([5], 100, 10, 10)).toBe('50,5');
  });
  it('clamps values above the maximum', () => {
    expect(sparklinePoints([200], 10, 10, 100)).toBe('5,0');
  });
});
