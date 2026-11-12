import type { ApiToken } from '../domain/apiToken';
import type {
  Passkey,
  PasskeyAssertion,
  PasskeyAttestation,
  PasskeyRegistrationOptions,
  SecondFactorChallenge,
} from '../domain/passkey';
import type { AuthStatus, Guest, Session } from '../domain/auth';
import type { SecurityEvent } from '../domain/securityEvent';
import type { SessionInfo } from '../domain/session';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';

/** Output port: talks to the authentication backend. */
export interface AuthGateway {
  status(): Promise<AuthStatus>;
  setup(setupToken: string, password: string): Promise<Session>;
  /** Returns a challenge instead of a session when a passkey is registered. */
  login(password: string, code?: string, username?: string): Promise<Session | SecondFactorChallenge>;
  passkeyLogin(ticket: string, assertion: PasskeyAssertion): Promise<Session>;
  passkeys(): Promise<{ available: boolean; passkeys: Passkey[] }>;
  beginPasskeyRegistration(): Promise<PasskeyRegistrationOptions>;
  finishPasskeyRegistration(name: string, attestation: PasskeyAttestation): Promise<string[]>;
  deletePasskey(id: string): Promise<void>;
  guests(): Promise<Guest[]>;
  createGuest(name: string, password: string, tabs: string[]): Promise<Guest>;
  updateGuest(id: string, tabs: string[], disabled: boolean, password: string): Promise<Guest>;
  deleteGuest(id: string): Promise<void>;
  apiTokens(): Promise<ApiToken[]>;
  createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string>;
  revokeApiToken(id: string): Promise<void>;
  /** Only accepted from a trusted reverse proxy that sets the user header. */
  headerLogin(): Promise<Session | { ticket: string }>;
  ssoSecondFactor(ticket: string, code: string): Promise<Session>;
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

/** Output port: the browser's WebAuthn API (isolated so use cases stay testable). */
export interface PasskeyAuthenticator {
  supported(): boolean;
  create(options: PasskeyRegistrationOptions): Promise<PasskeyAttestation>;
  get(challenge: SecondFactorChallenge): Promise<PasskeyAssertion>;
}
