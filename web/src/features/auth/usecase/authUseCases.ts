import { type AuthView, validateNewPassword, viewFor } from '../domain/auth';
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
    if (setupToken.trim() === '')
      throw new ValidationError('Please enter the setup token from the server log.');
    const session = await this.gateway.setup(setupToken.trim(), password);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /** Logs in with the password. */
  async login(password: string): Promise<void> {
    if (password === '') throw new ValidationError('Please enter your password.');
    const session = await this.gateway.login(password);
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
}
