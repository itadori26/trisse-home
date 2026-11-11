import type { MessageKey } from '../../../shared/i18n';

/** One entry of the security log. */
export interface SecurityEvent {
  id: string;
  at: Date;
  type: string;
  ip: string;
  userAgent: string;
  detail: string;
}

const labels: Record<string, MessageKey> = {
  login_succeeded: 'auth.event.login_succeeded',
  login_failed: 'auth.event.login_failed',
  login_locked: 'auth.event.login_locked',
  new_device: 'auth.event.new_device',
  logout_everywhere: 'auth.event.logout_everywhere',
  password_changed: 'auth.event.password_changed',
  two_factor_enabled: 'auth.event.two_factor_enabled',
  two_factor_disabled: 'auth.event.two_factor_disabled',
  recovery_code_used: 'auth.event.recovery_code_used',
  session_revoked: 'auth.event.session_revoked',
  external: 'auth.event.external',
};

/** Message key for an event type; unknown types fall back to a generic label. */
export function eventLabel(type: string): MessageKey {
  return labels[type] ?? 'auth.event.other';
}

/** Events that deserve attention are highlighted in the list. */
export function isWarning(type: string): boolean {
  return (
    type === 'login_failed' ||
    type === 'login_locked' ||
    type === 'new_device' ||
    type === 'recovery_code_used'
  );
}
