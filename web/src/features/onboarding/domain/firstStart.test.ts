import { describe, expect, it } from 'vitest';
import { firstStartSteps } from './firstStart';

describe('firstStartSteps (#671)', () => {
  const all = { owner: true, personal: true, look: true, places: true, extra: ['instance'] };

  it('walks the owner through every step in order', () => {
    expect(firstStartSteps(all)).toEqual(['name', 'place', 'language', 'look', 'extra:instance']);
  });

  it('asks invited members the personal steps only', () => {
    expect(firstStartSteps({ ...all, owner: false })).toEqual(['name', 'language', 'look']);
  });

  it('leaves out steps whose setting is not available', () => {
    expect(firstStartSteps({ ...all, look: false, places: false, extra: [] })).toEqual(['name', 'language']);
    expect(firstStartSteps({ owner: true, personal: false, look: false, places: false, extra: [] })).toEqual(
      [],
    );
  });
});
