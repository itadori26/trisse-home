import type { AuthStatus, Session } from '../domain/auth';

/** Output port: talks to the authentication backend. */
export interface AuthGateway {
  status(): Promise<AuthStatus>;
  setup(setupToken: string, password: string): Promise<Session>;
  login(password: string): Promise<Session>;
  logout(): Promise<void>;
  changePassword(currentPassword: string, newPassword: string): Promise<Session>;
}

/** Output port: receives the CSRF token so later requests can send it. */
export interface CsrfTokenStore {
  setCsrfToken(token: string): void;
}
