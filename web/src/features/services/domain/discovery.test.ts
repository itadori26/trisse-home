import { describe, expect, it } from 'vitest';
import { labelExample, MAX_INTERVAL, validInterval } from './discovery';
import { visibleServices } from './service';

describe('discovery', () => {
  it('accepts whole minutes within the limits', () => {
    expect(validInterval(1)).toBe(true);
    expect(validInterval(MAX_INTERVAL)).toBe(true);
    expect(validInterval(0)).toBe(false);
    expect(validInterval(MAX_INTERVAL + 1)).toBe(false);
    expect(validInterval(1.5)).toBe(false);
    expect(validInterval(Number.NaN)).toBe(false);
  });
  it('shows the labels the backend reads', () => {
    for (const key of ['name', 'url', 'icon', 'description'])
      expect(labelExample).toContain(`trisse-home.${key}:`);
  });
  it('filters hidden tiles', () => {
    expect(visibleServices([{ hidden: false }, { hidden: true }])).toEqual([{ hidden: false }]);
  });
});
