import { t } from '../../../shared/i18n';

/** Something dropped to pick up on another device (#101). */
export interface DropItem {
  id: string;
  kind: 'text' | 'link' | 'file';
  text: string;
  fileName: string;
  contentType: string;
  size: number;
  createdAt: Date;
  expiresAt: Date;
}

/** Expiries offered, in hours. */
export const EXPIRIES = [1, 24, 168] as const;
export type ExpiryHours = (typeof EXPIRIES)[number];
export const MAX_FILE_BYTES = 25 << 20;

/** "in 3 h", "in 5 days": time until an item expires. */
export function expiresIn(item: Pick<DropItem, 'expiresAt'>, now: Date): string {
  const minutes = Math.max(0, Math.round((item.expiresAt.getTime() - now.getTime()) / 60_000));
  if (minutes < 60) return t('drop.in_minutes', { n: minutes });
  const hours = Math.round(minutes / 60);
  if (hours < 48) return t('drop.in_hours', { n: hours });
  return t('drop.in_days', { n: Math.round(hours / 24) });
}

/** Checks a file before uploading it (the server checks again). */
export function validateFile(f: { size: number }): string | null {
  if (f.size === 0) return t('drop.file_empty');
  if (f.size > MAX_FILE_BYTES) return t('drop.file_too_large');
  return null;
}

/** "1.2 MB" style sizes. */
export function formatSize(n: number, locale: string): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toLocaleString(locale, { maximumFractionDigits: 0 })} KB`;
  return `${(n / 1024 / 1024).toLocaleString(locale, { maximumFractionDigits: 1 })} MB`;
}
