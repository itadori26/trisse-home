import { t } from '../../../shared/i18n';
/** Minimum password length; mirrors the backend password policy. */
export const MIN_PASSWORD_LENGTH = 12;

/** Authentication status of the current browser. */
export interface AuthStatus {
  setupRequired: boolean;
  authenticated: boolean;
  csrfToken: string | null;
  /** Public demo instance: signs in without a password. */
  demo: boolean;
  /** Read-only guest account (ADR 0012); also true for kiosk displays. */
  guest: boolean;
  /** Kiosk display (#50): no controls, rotation, night dimming. */
  kiosk: boolean;
  /** Available sign-in methods (ADR 0013). */
  sso: { oidc: boolean; header: boolean; localLogin: boolean };
}

/** Result of a single sign-on attempt. */
export type SsoResult = { kind: 'signed-in' } | { kind: 'needs-code'; ticket: string };

/** Message key for an error the provider redirect reported (?sso_error=…). */
export function ssoErrorKey(
  code: string | null,
): 'auth.sso_error.not_allowed' | 'auth.sso_error.failed' | null {
  if (!code) return null;
  return code === 'not_allowed' ? 'auth.sso_error.not_allowed' : 'auth.sso_error.failed';
}

/** A read-only guest account as managed by the owner. */
export interface Guest {
  id: string;
  name: string;
  tabs: string[];
  disabled: boolean;
  createdAt: Date;
}

/** Guest names: 2–32 lowercase letters, digits, dots, dashes, underscores. */
export function validGuestName(name: string): boolean {
  return /^[a-z0-9][a-z0-9._-]{1,31}$/.test(name) && name !== 'owner';
}

/** A freshly started session. */
export interface Session {
  csrfToken: string;
  expiresAt: Date;
}

/** Which screen the app has to show. */
export type AuthView = 'setup' | 'login' | 'app';

/** Derives the screen to show from the authentication status. */
export function viewFor(status: AuthStatus): AuthView {
  if (status.setupRequired) return 'setup';
  return status.authenticated ? 'app' : 'login';
}

/**
 * Validates a new password and its confirmation.
 * @returns an error message, or null if the password is acceptable
 */
export function validateNewPassword(password: string, confirmation: string): string | null {
  if ([...password].length < MIN_PASSWORD_LENGTH) {
    return t('auth.password_too_short', { min: MIN_PASSWORD_LENGTH });
  }
  if (password.trim() === '') return t('auth.password_whitespace');
  if (password !== confirmation) return t('auth.passwords_differ');
  return null;
}

/** How a kiosk display shows its tabs (#50). */
export interface KioskOptions {
  /** Seconds per tab; 0 = no rotation. */
  rotateSeconds: number;
  /** "HH:MM" local time; both empty = never dimmed. */
  nightFrom: string;
  nightTo: string;
}

/** A kiosk display as managed by the owner. */
export interface KioskDisplay extends KioskOptions {
  id: string;
  name: string;
  tabs: string[];
  disabled: boolean;
  createdAt: Date;
  lastUsedAt: Date | null;
}

/** Input of creating or changing a display. */
export interface KioskInput extends KioskOptions {
  name: string;
  tabs: string[];
  disabled: boolean;
}

/** The link that signs a display in; the token travels in the fragment, never to the server log. */
export function kioskLink(origin: string, token: string): string {
  return `${origin}/kiosk#${encodeURIComponent(token)}`;
}

/** Whether a "HH:MM" time window (possibly over midnight) contains now. */
export function inNightWindow(now: Date, from: string, to: string): boolean {
  if (!from || !to) return false;
  const minutes = (s: string) => Number(s.slice(0, 2)) * 60 + Number(s.slice(3, 5));
  const m = now.getHours() * 60 + now.getMinutes();
  const a = minutes(from);
  const b = minutes(to);
  return a < b ? m >= a && m < b : m >= a || m < b;
}
