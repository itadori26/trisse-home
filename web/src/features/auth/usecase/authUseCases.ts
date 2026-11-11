import { t } from '../../../shared/i18n';
import { type AuthView, validateNewPassword, viewFor } from '../domain/auth';
import type { SessionInfo } from '../domain/session';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';
import type { AuthGateway, CsrfTokenStore } from './ports';

/** Error for input rejected before it reaches the backend. */
export class ValidationError extends Error {
  override name = 'ValidationError';
}

/** Application logic of the auth feature. */
export class AuthUseCases {
  constructor(
    private readonly gateway: AuthGateway,
    private readonly csrf: CsrfTokenStore,
  ) {}

  /** Loads the status and returns the screen to show. */
  async currentView(): Promise<AuthView> {
    const status = await this.gateway.status();
    if (status.csrfToken) this.csrf.setCsrfToken(status.csrfToken);
    return viewFor(status);
  }

  /** Creates the account with the setup token from the server log. */
  async setup(setupToken: string, password: string, confirmation: string): Promise<void> {
    const problem = validateNewPassword(password, confirmation);
    if (problem) throw new ValidationError(problem);
    if (setupToken.trim() === '') throw new ValidationError(t('auth.enter_setup_token'));
    const session = await this.gateway.setup(setupToken.trim(), password);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /** Logs in with the password and, if two-factor authentication is enabled, a code. */
  async login(password: string, code?: string): Promise<void> {
    if (password === '') throw new ValidationError(t('auth.enter_password'));
    const session = await this.gateway.login(password, code?.trim() || undefined);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /** Ends the current session. */
  async logout(): Promise<void> {
    await this.gateway.logout();
    this.csrf.setCsrfToken('');
  }

  /** Changes the password; all other sessions are signed out by the backend. */
  async changePassword(current: string, next: string, confirmation: string): Promise<void> {
    const problem = validateNewPassword(next, confirmation);
    if (problem) throw new ValidationError(problem);
    const session = await this.gateway.changePassword(current, next);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /** Lists active sessions; the current one first. */
  async sessions(): Promise<SessionInfo[]> {
    const list = await this.gateway.sessions();
    return [...list].sort(
      (a, b) => Number(b.current) - Number(a.current) || b.lastSeenAt.getTime() - a.lastSeenAt.getTime(),
    );
  }

  revokeSession(id: string): Promise<void> {
    return this.gateway.revokeSession(id);
  }

  /** Signs out every other device; returns how many sessions ended. */
  revokeOtherSessions(): Promise<number> {
    return this.gateway.revokeOtherSessions();
  }

  twoFactorStatus(): Promise<TwoFactorStatus> {
    return this.gateway.twoFactorStatus();
  }

  /** Starts connecting an authenticator app; 2FA is active only after confirmation. */
  startTwoFactor(): Promise<TwoFactorEnrollment> {
    return this.gateway.startTwoFactor();
  }

  /** Confirms the setup with a code and returns the one-time recovery codes. */
  confirmTwoFactor(code: string): Promise<string[]> {
    const c = code.replace(/\s/g, '');
    if (!/^\d{6}$/.test(c)) return Promise.reject(new ValidationError(t('auth.enter_code')));
    return this.gateway.confirmTwoFactor(c);
  }

  disableTwoFactor(password: string, code: string): Promise<void> {
    if (password === '' || code.trim() === '') {
      return Promise.reject(new ValidationError(t('auth.password_and_code_required')));
    }
    return this.gateway.disableTwoFactor(password, code.trim());
  }
}
