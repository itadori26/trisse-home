/** Minimum password length; mirrors the backend password policy. */
export const MIN_PASSWORD_LENGTH = 12;

/** Authentication status of the current browser. */
export interface AuthStatus {
  setupRequired: boolean;
  authenticated: boolean;
  csrfToken: string | null;
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
    return `The password must be at least ${MIN_PASSWORD_LENGTH} characters long.`;
  }
  if (password.trim() === '') return 'The password must not consist of whitespace only.';
  if (password !== confirmation) return 'The passwords do not match.';
  return null;
}
