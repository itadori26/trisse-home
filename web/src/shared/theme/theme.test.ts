import { describe, expect, it } from 'vitest';
import { resolveTheme } from './theme';

describe('resolveTheme', () => {
  it.each([
    ['system', true, 'dark'],
    ['system', false, 'light'],
    ['light', true, 'light'],
    ['dark', false, 'dark'],
  ] as const)('%s with prefersDark=%s -> %s', (mode, prefersDark, want) => {
    expect(resolveTheme(mode, prefersDark)).toBe(want);
  });
});
