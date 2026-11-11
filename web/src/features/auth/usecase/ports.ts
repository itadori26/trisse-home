import type { ApiToken } from '../domain/apiToken';
import type { AuthStatus, Guest, Session } from '../domain/auth';
import type { SecurityEvent } from '../domain/securityEvent';
import type { SessionInfo } from '../domain/session';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';

/** Output port: talks to the authentication backend. */
export interface AuthGateway {
  status(): Promise<AuthStatus>;
  setup(setupToken: string, password: string): Promise<Session>;
  login(password: string, code?: string, username?: string): Promise<Session>;
  guests(): Promise<Guest[]>;
  createGuest(name: string, password: string, tabs: string[]): Promise<Guest>;
  updateGuest(id: string, tabs: string[], disabled: boolean, password: string): Promise<Guest>;
  deleteGuest(id: string): Promise<void>;
  apiTokens(): Promise<ApiToken[]>;
  createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string>;
  revokeApiToken(id: string): Promise<void>;
  /** Only available on a demo instance. */
  demoLogin(): Promise<Session>;
  logout(): Promise<void>;
  changePassword(currentPassword: string, newPassword: string): Promise<Session>;
  sessions(): Promise<SessionInfo[]>;
  securityEvents(): Promise<SecurityEvent[]>;
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
