import type { AuthStatus, Session } from '../domain/auth';
import type { SessionInfo } from '../domain/session';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';

/** Output port: talks to the authentication backend. */
export interface AuthGateway {
  status(): Promise<AuthStatus>;
  setup(setupToken: string, password: string): Promise<Session>;
  login(password: string, code?: string): Promise<Session>;
  logout(): Promise<void>;
  changePassword(currentPassword: string, newPassword: string): Promise<Session>;
  sessions(): Promise<SessionInfo[]>;
  revokeSession(id: string): Promise<void>;
  revokeOtherSessions(): Promise<number>;
  twoFactorStatus(): Promise<TwoFactorStatus>;
  startTwoFactor(): Promise<TwoFactorEnrollment>;
  confirmTwoFactor(code: string): Promise<string[]>;
  disableTwoFactor(password: string, code: string): Promise<void>;
}

/** Output port: receives the CSRF token so later requests can send it. */
export interface CsrfTokenStore {
  setCsrfToken(token: string): void;
}
