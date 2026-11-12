import { t } from '../../../shared/i18n';
import type { ApiToken } from '../domain/apiToken';
import { type AuthView, type Guest, validGuestName, validateNewPassword, viewFor } from '../domain/auth';
import type { SecurityEvent } from '../domain/securityEvent';
import type { SessionInfo } from '../domain/session';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';
import type { Passkey, SecondFactorChallenge } from '../domain/passkey';
import type { AuthGateway, CsrfTokenStore, PasskeyAuthenticator } from './ports';

/** Error for input rejected before it reaches the backend. */
export class ValidationError extends Error {
  override name = 'ValidationError';
}

/** Application logic of the auth feature. */
export class AuthUseCases {
  constructor(
    private readonly gateway: AuthGateway,
    private readonly csrf: CsrfTokenStore,
    private readonly authenticator?: PasskeyAuthenticator,
  ) {}

  /** Whether this browser can use passkeys here (secure context). */
  passkeysSupported(): boolean {
    return this.authenticator?.supported() ?? false;
  }

  /** True on a public demo instance (set by currentView). */
  demo = false;
  /** True when a read-only guest is signed in (set by currentView). */
  guest = false;
  /** Sign-in methods offered by the server (set by currentView). */
  sso = { oidc: false, header: false, localLogin: true };
  /** Ticket of an SSO sign-in that still needs the local code. */
  ssoTicket = '';

  /** Loads the status and returns the screen to show; a demo signs in automatically. */
  async currentView(): Promise<AuthView> {
    const status = await this.gateway.status();
    this.demo = status.demo;
    this.guest = status.guest;
    this.sso = status.sso;
    // Behind a reverse proxy with forward auth the user is already known.
    if (status.sso.header && !status.authenticated && !status.setupRequired) {
      try {
        const res = await this.gateway.headerLogin();
        if ('ticket' in res) this.ssoTicket = res.ticket;
        else {
          this.csrf.setCsrfToken(res.csrfToken);
          return 'app';
        }
      } catch {
        // Not allowed or not configured: fall back to the login page.
      }
    }
    if (status.csrfToken) this.csrf.setCsrfToken(status.csrfToken);
    if (status.demo && !status.authenticated && !status.setupRequired) {
      const session = await this.gateway.demoLogin();
      this.csrf.setCsrfToken(session.csrfToken);
      return 'app';
    }
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

  /**
   * Logs in with the password and, if two-factor authentication is enabled,
   * a code. Guests also give their name.
   */
  async login(password: string, code?: string, username?: string): Promise<SecondFactorChallenge | null> {
    if (password === '') throw new ValidationError(t('auth.enter_password'));
    const name = username?.trim().toLowerCase() || undefined;
    const res = await this.gateway.login(password, code?.trim() || undefined, name);
    if ('ticket' in res) return res;
    this.csrf.setCsrfToken(res.csrfToken);
    return null;
  }

  /** Second login step: sign the challenge with a passkey. */
  async loginWithPasskey(challenge: SecondFactorChallenge): Promise<void> {
    if (!this.authenticator) throw new ValidationError(t('auth.passkeys_unsupported'));
    const assertion = await this.authenticator.get(challenge);
    const session = await this.gateway.passkeyLogin(challenge.ticket, assertion);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  passkeys(): Promise<{ available: boolean; passkeys: Passkey[] }> {
    return this.gateway.passkeys();
  }

  /** Registers a passkey; returns new recovery codes when they were issued. */
  async addPasskey(name: string): Promise<string[]> {
    if (!this.authenticator?.supported()) throw new ValidationError(t('auth.passkeys_unsupported'));
    const options = await this.gateway.beginPasskeyRegistration();
    const attestation = await this.authenticator.create(options);
    return this.gateway.finishPasskeyRegistration(name.trim(), attestation);
  }

  deletePasskey(id: string): Promise<void> {
    return this.gateway.deletePasskey(id);
  }

  /** Finishes an SSO sign-in that needs the local authentication code. */
  async completeSso(ticket: string, code: string): Promise<void> {
    if (code.trim() === '') throw new ValidationError(t('auth.enter_code_hint'));
    const session = await this.gateway.ssoSecondFactor(ticket, code.trim());
    this.ssoTicket = '';
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

  /** Guest accounts (owner only). */
  guests(): Promise<Guest[]> {
    return this.gateway.guests();
  }

  /** Creates a read-only guest that sees the given tabs. */
  async createGuest(name: string, password: string, confirmation: string, tabs: string[]): Promise<Guest> {
    const n = name.trim().toLowerCase();
    if (!validGuestName(n)) throw new ValidationError(t('auth.guest_name_invalid'));
    const problem = validateNewPassword(password, confirmation);
    if (problem) throw new ValidationError(problem);
    return this.gateway.createGuest(n, password, tabs);
  }

  /** Changes shared tabs, the disabled flag and optionally the password. */
  async updateGuest(guest: Guest, tabs: string[], disabled: boolean, password = ''): Promise<Guest> {
    if (password !== '') {
      const problem = validateNewPassword(password, password);
      if (problem) throw new ValidationError(problem);
    }
    return this.gateway.updateGuest(guest.id, tabs, disabled, password);
  }

  deleteGuest(id: string): Promise<void> {
    return this.gateway.deleteGuest(id);
  }

  /** Personal API tokens (owner only). */
  apiTokens(): Promise<ApiToken[]> {
    return this.gateway.apiTokens();
  }

  /** Creates a token and returns its secret, which is shown only once. */
  async createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string> {
    if (name.trim() === '') throw new ValidationError(t('auth.token_name_required'));
    if (scopes.length === 0) throw new ValidationError(t('auth.token_scope_required'));
    return this.gateway.createApiToken(name.trim(), scopes, Math.max(0, Math.round(expiresInDays) || 0));
  }

  revokeApiToken(id: string): Promise<void> {
    return this.gateway.revokeApiToken(id);
  }

  /** The security log, newest first. */
  securityEvents(): Promise<SecurityEvent[]> {
    return this.gateway.securityEvents();
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
