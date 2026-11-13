import { describe, expect, it } from 'vitest';
import { dataAge } from './answer';

describe('dataAge', () => {
  const now = new Date(2026, 10, 12, 12, 0);
  const ago = (min: number) => new Date(now.getTime() - min * 60_000);
  it('says how old the data is', () => {
    expect(dataAge(null, now)).toBeNull();
    expect(dataAge(ago(0.5), now)).toEqual({ key: 'smart.age_now', n: 0 });
    expect(dataAge(ago(3), now)).toEqual({ key: 'smart.age_min', n: 3 });
    expect(dataAge(ago(119), now)).toEqual({ key: 'smart.age_min', n: 119 });
    expect(dataAge(ago(180), now)).toEqual({ key: 'smart.age_hours', n: 3 });
    expect(dataAge(new Date(now.getTime() + 5000), now)).toEqual({ key: 'smart.age_now', n: 0 });
  });
});
