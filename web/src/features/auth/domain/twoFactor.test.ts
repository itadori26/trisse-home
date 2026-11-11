import { describe, expect, it } from 'vitest';
import { formatSecret, looksLikeCode, recoveryCodesText } from './twoFactor';

describe('two-factor domain', () => {
  it('groups the secret', () => expect(formatSecret('JBSWY3DPEHPK3PXP')).toBe('JBSW Y3DP EHPK 3PXP'));
  it('recognises codes', () => {
    expect(looksLikeCode('123 456')).toBe(true);
    expect(looksLikeCode('abcd-efgh')).toBe(true);
    expect(looksLikeCode('12345')).toBe(false);
  });
  it('builds the recovery code file', () => {
    expect(recoveryCodesText(['abcd-efgh'], new Date('2026-11-11T00:00:00Z'))).toContain('abcd-efgh');
  });
});
