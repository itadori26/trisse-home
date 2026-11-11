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
