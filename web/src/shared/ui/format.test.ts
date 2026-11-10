import { describe, expect, it } from 'vitest';
import { formatBytes, formatCount, formatUptime } from './format';

describe('format', () => {
  it.each([
    [512, '512 B'],
    [1536, '1.5 KiB'],
    [1024 ** 3 * 12.4, '12 GiB'],
  ])('formatBytes(%d) = %s', (n, want) => expect(formatBytes(n)).toBe(want));
  it.each([
    [59, '0m'],
    [3700, '1h 1m'],
    [90000, '1d 1h'],
  ])('formatUptime(%d) = %s', (n, want) => expect(formatUptime(n)).toBe(want));
  it('formats counts', () => expect(formatCount(12345)).toBe('12.3K'));
});
