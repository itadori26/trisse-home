import { describe, expect, it } from 'vitest';
import {
  clampRect,
  collides,
  compact,
  describeRect,
  findFreeSpot,
  keyboardStep,
  moveItem,
  resizeItem,
  sameLayout,
  type LayoutItem,
} from './layout';

const item = (id: string, x: number, y: number, w: number, h: number): LayoutItem => ({ id, x, y, w, h });
const noOverlaps = (items: LayoutItem[]) =>
  items.every((a, i) => items.every((b, j) => i === j || !collides(a, b)));

describe('collides', () => {
  it('detects overlap but not adjacency', () => {
    expect(collides(item('a', 0, 0, 2, 2), item('b', 1, 1, 2, 2))).toBe(true);
    expect(collides(item('a', 0, 0, 2, 2), item('b', 2, 0, 2, 2))).toBe(false);
    expect(collides(item('a', 0, 0, 2, 2), item('b', 0, 2, 2, 2))).toBe(false);
  });
});

describe('clampRect', () => {
  it('keeps rectangles inside the grid', () => {
    expect(clampRect({ x: 11, y: -3, w: 4, h: 0 })).toEqual({ x: 8, y: 0, w: 4, h: 1 });
    expect(clampRect({ x: 0, y: 0, w: 20, h: 99 })).toEqual({ x: 0, y: 0, w: 12, h: 24 });
  });
});

describe('compact', () => {
  it('pulls items up without overlaps', () => {
    const out = compact([item('a', 0, 5, 4, 2), item('b', 0, 9, 4, 2), item('c', 4, 3, 4, 1)]);
    expect(out.find((i) => i.id === 'a')).toMatchObject({ y: 0 });
    expect(out.find((i) => i.id === 'b')).toMatchObject({ y: 2 });
    expect(out.find((i) => i.id === 'c')).toMatchObject({ y: 0 });
  });
  it('keeps the pinned item in place', () => {
    const out = compact([item('a', 0, 0, 4, 2), item('p', 0, 1, 4, 2)], 'p');
    expect(out.find((i) => i.id === 'p')).toMatchObject({ y: 1 });
    expect(noOverlaps(out)).toBe(true);
  });
});

describe('moveItem', () => {
  it('swaps vertical order when dragging an item above another', () => {
    const items = [item('a', 0, 0, 6, 2), item('b', 0, 2, 6, 2)];
    const out = moveItem(items, 'b', 0, 0);
    expect(out.find((i) => i.id === 'b')).toMatchObject({ y: 0 });
    expect(out.find((i) => i.id === 'a')).toMatchObject({ y: 2 });
  });
  it('moves an item to a free column', () => {
    const out = moveItem([item('a', 0, 0, 4, 2), item('b', 0, 2, 4, 2)], 'b', 6, 0);
    expect(out.find((i) => i.id === 'b')).toMatchObject({ x: 6, y: 0 });
    expect(noOverlaps(out)).toBe(true);
  });
  it('clamps targets outside the grid', () => {
    const out = moveItem([item('a', 0, 0, 4, 2)], 'a', 20, -5);
    expect(out[0]).toMatchObject({ x: 8, y: 0 });
  });
});

describe('resizeItem', () => {
  it('pushes items below when growing', () => {
    const out = resizeItem([item('a', 0, 0, 4, 2), item('b', 0, 2, 4, 2)], 'a', 4, 5);
    expect(out.find((i) => i.id === 'b')).toMatchObject({ y: 5 });
    expect(noOverlaps(out)).toBe(true);
  });
});

describe('findFreeSpot', () => {
  it('fills gaps before adding rows', () => {
    expect(findFreeSpot([item('a', 0, 0, 6, 2)], 6, 2)).toEqual({ x: 6, y: 0, w: 6, h: 2 });
    expect(findFreeSpot([item('a', 0, 0, 12, 2)], 3, 3)).toEqual({ x: 0, y: 2, w: 3, h: 3 });
  });
});

describe('sameLayout', () => {
  it('compares by id and rectangle', () => {
    const a = [item('a', 0, 0, 1, 1), item('b', 1, 0, 1, 1)];
    expect(sameLayout(a, [...a].reverse())).toBe(true);
    expect(sameLayout(a, [item('a', 0, 1, 1, 1), item('b', 1, 0, 1, 1)])).toBe(false);
  });

  describe('keyboardStep', () => {
    const items = [
      { id: 'a', x: 0, y: 0, w: 4, h: 2 },
      { id: 'b', x: 0, y: 2, w: 4, h: 2 },
    ];
    it('moves right and stops at the edge', () => {
      expect(keyboardStep(items, 'a', 'ArrowRight', false)?.find((i) => i.id === 'a')?.x).toBe(1);
      const edge = [{ id: 'a', x: 8, y: 0, w: 4, h: 2 }];
      expect(keyboardStep(edge, 'a', 'ArrowRight', false)).toBeNull();
    });
    it('moves down past the widget below', () => {
      const next = keyboardStep(items, 'a', 'ArrowDown', false);
      const a = next?.find((i) => i.id === 'a');
      const b = next?.find((i) => i.id === 'b');
      expect(b?.y).toBe(0);
      expect(a?.y).toBe(2);
    });
    it('resizes with shift', () => {
      expect(keyboardStep(items, 'a', 'ArrowRight', true)?.find((i) => i.id === 'a')?.w).toBe(5);
      expect(keyboardStep([{ id: 'a', x: 0, y: 0, w: 1, h: 1 }], 'a', 'ArrowLeft', true)).toBeNull();
    });
    it('describes positions', () => {
      expect(describeRect({ x: 0, y: 2, w: 4, h: 3 })).toBe('column 1, row 3, 4 × 3 cells');
    });
  });
});
