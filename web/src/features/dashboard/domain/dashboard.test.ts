import { describe, expect, it } from 'vitest';
import { reorder, validateBackground, validateTabName } from './dashboard';

describe('validateTabName', () => {
  it.each([
    ['', /empty/],
    ['   ', /empty/],
    ['x'.repeat(41), /40/],
  ])('rejects %j', (name, msg) => expect(validateTabName(name)).toMatch(msg));
  it('accepts normal names', () => expect(validateTabName(' Server ')).toBeNull());
});

describe('reorder', () => {
  it('moves elements', () => {
    expect(reorder(['a', 'b', 'c'], 0, 2)).toEqual(['b', 'c', 'a']);
    expect(reorder(['a', 'b', 'c'], 2, 0)).toEqual(['c', 'a', 'b']);
  });

  it('validates tab backgrounds', () => {
    expect(validateBackground('')).toBeNull();
    expect(validateBackground('https://img.example/a.jpg')).toBeNull();
    expect(validateBackground('javascript:alert(1)')).toMatch(/http/);
  });
});
