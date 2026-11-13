import { describe, expect, it } from 'vitest';
import { deriveMobile, moveMobile, resolveMobile, updateMobile } from './mobile';

const w = (id: string, x: number, y: number, h = 2) => ({ id, x, y, w: 4, h });

describe('phone layout (#92)', () => {
  it('derives reading order and desktop heights', () => {
    expect(deriveMobile([w('c', 0, 4), w('b', 8, 0, 3), w('a', 0, 0, 5)])).toEqual([
      { id: 'a', hidden: false, rows: 5 },
      { id: 'b', hidden: false, rows: 3 },
      { id: 'c', hidden: false, rows: 2 },
    ]);
  });

  it('keeps the stored order, drops deleted and appends new widgets', () => {
    const stored = [
      { id: 'b', hidden: true, rows: 99 },
      { id: 'gone', hidden: false, rows: 2 },
      { id: 'a', hidden: false, rows: 1 },
      { id: 'a', hidden: false, rows: 7 },
    ];
    expect(resolveMobile(stored, [w('a', 0, 0), w('b', 4, 0), w('new', 0, 9, 3)])).toEqual([
      { id: 'b', hidden: true, rows: 24 },
      { id: 'a', hidden: false, rows: 1 },
      { id: 'new', hidden: false, rows: 3 },
    ]);
  });

  it('moves, hides and resizes items', () => {
    const items = deriveMobile([w('a', 0, 0), w('b', 0, 2), w('c', 0, 4)]);
    expect(moveMobile(items, 'c', -2)?.map((i) => i.id)).toEqual(['c', 'a', 'b']);
    expect(moveMobile(items, 'a', -1)).toBeNull();
    expect(moveMobile(items, 'c', 1)).toBeNull();
    expect(updateMobile(items, 'b', { hidden: true })[1]).toEqual({ id: 'b', hidden: true, rows: 2 });
    expect(updateMobile(items, 'b', { rows: 0 })[1]?.rows).toBe(1);
    expect(updateMobile(items, 'b', { rows: 30 })[1]?.rows).toBe(24);
  });
});
