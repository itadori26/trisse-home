import { describe, expect, it } from 'vitest';
import { fromBase64Url, toBase64Url } from './passkey';

describe('base64url', () => {
  it('round-trips bytes without padding', () => {
    const bytes = new Uint8Array([0, 251, 255, 1, 2]);
    const s = toBase64Url(bytes);
    expect(s).not.toMatch(/[+/=]/);
    expect([...fromBase64Url(s)]).toEqual([...bytes]);
  });
});
