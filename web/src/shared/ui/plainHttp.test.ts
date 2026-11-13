import { describe, expect, it } from 'vitest';
import { usesPlainHttp } from './plainHttp';

describe('usesPlainHttp', () => {
  it('warns for http:// in any line, not for https or empty input', () => {
    expect(usesPlainHttp('http://192.168.1.2:3000')).toBe(true);
    expect(usesPlainHttp(' HTTP://nas.local')).toBe(true);
    expect(usesPlainHttp('https://a.example\nhttp://b.lan/dav/')).toBe(true);
    expect(usesPlainHttp('https://nas.local')).toBe(false);
    expect(usesPlainHttp('')).toBe(false);
    expect(usesPlainHttp('nas.local/http://')).toBe(false);
  });
});
