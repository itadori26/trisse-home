import { describe, expect, it } from 'vitest';
import { shortVersion } from './nextcloud';

describe('shortVersion', () => {
  it('drops build suffixes', () => {
    expect(shortVersion('30.0.2')).toBe('30.0.2');
    expect(shortVersion('29.0.8 RC1')).toBe('29.0.8');
  });
});
