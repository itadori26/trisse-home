import { t } from '../../../shared/i18n';
/** Two-factor authentication state. */
export interface TwoFactorStatus {
  enabled: boolean;
  recoveryCodesLeft: number;
}

/** Data shown while connecting an authenticator app. */
export interface TwoFactorEnrollment {
  secret: string;
  uri: string;
}

/** Groups the base32 secret for manual entry, e.g. "JBSW Y3DP …". */
export function formatSecret(secret: string): string {
  return secret.replace(/(.{4})/g, '$1 ').trim();
}

/** Text file content for downloading the recovery codes. */
export function recoveryCodesText(codes: string[], createdAt: Date): string {
  return [
    t('auth.recovery_file_title'),
    t('auth.recovery_file_created', { date: createdAt.toISOString() }),
    t('auth.recovery_file_hint'),
    '',
    ...codes,
    '',
  ].join('\n');
}

/** Validates a 6-digit authenticator code or a recovery code. */
export function looksLikeCode(input: string): boolean {
  const s = input.replace(/[\s-]/g, '');
  return /^\d{6}$/.test(s) || /^[a-z2-7]{8}$/i.test(s);
}
