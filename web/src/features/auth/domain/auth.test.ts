import { describe, expect, it } from 'vitest';
import { validateNewPassword, viewFor } from './auth';

describe('viewFor', () => {
  it('prefers setup over everything else', () => {
    expect(
      viewFor({
        setupRequired: true,
        authenticated: false,
        csrfToken: null,
        demo: false,
        guest: false,
        kiosk: false,
        sso: { oidc: false, header: false, localLogin: true },
      }),
    ).toBe('setup');
  });
  it('shows the app for authenticated users', () => {
    expect(
      viewFor({
        setupRequired: false,
        authenticated: true,
        csrfToken: 'x',
        demo: false,
        guest: false,
        kiosk: false,
        sso: { oidc: false, header: false, localLogin: true },
      }),
    ).toBe('app');
  });
  it('shows the login otherwise', () => {
    expect(
      viewFor({
        setupRequired: false,
        authenticated: false,
        csrfToken: null,
        demo: false,
        guest: false,
        kiosk: false,
        sso: { oidc: false, header: false, localLogin: true },
      }),
    ).toBe('login');
  });
});

describe('validateNewPassword', () => {
  it.each([
    ['short', 'short', /at least/],
    ['            ', '            ', /whitespace/],
    ['long-enough-password', 'different-password', /match/],
  ])('rejects %j', (pw, confirm, want) => {
    expect(validateNewPassword(pw, confirm)).toMatch(want);
  });
  it('accepts a good password', () => {
    expect(validateNewPassword('long-enough-password', 'long-enough-password')).toBeNull();
  });
});

describe('kiosk displays (#50)', () => {
  it('puts the token into the fragment and knows night windows over midnight', async () => {
    const { inNightWindow, kioskLink } = await import('./auth');
    expect(kioskLink('https://home.example', 'a/b')).toBe('https://home.example/kiosk#a%2Fb');
    const at = (h: number, m = 0) => new Date(2026, 10, 14, h, m);
    expect(inNightWindow(at(23), '22:00', '06:30')).toBe(true);
    expect(inNightWindow(at(6, 15), '22:00', '06:30')).toBe(true);
    expect(inNightWindow(at(6, 30), '22:00', '06:30')).toBe(false);
    expect(inNightWindow(at(13), '12:00', '14:00')).toBe(true);
    expect(inNightWindow(at(13), '', '')).toBe(false);
  });
});
