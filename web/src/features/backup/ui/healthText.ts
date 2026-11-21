import { intlLocale, t } from '../../../shared/i18n';
import { type BackupStatus, type HealthTone, dayWord, healthLine } from '../domain/backup';

/** Formats a moment as "today 03:00", "yesterday 03:00" or "21 Nov, 03:00". */
export function whenText(at: Date, now: Date): string {
  const time = new Intl.DateTimeFormat(intlLocale(), { hour: '2-digit', minute: '2-digit' }).format(at);
  const word = dayWord(at, now);
  if (word !== 'earlier') return t(`backup.when.${word}`, { time });
  const date = new Intl.DateTimeFormat(intlLocale(), { day: 'numeric', month: 'short' }).format(at);
  return t('backup.when.earlier', { date, time });
}

/** The traffic light in words, e.g. "Last backup today 03:00, all good". */
export function healthText(s: BackupStatus, now = new Date()): { tone: HealthTone; text: string } {
  const line = healthLine(s);
  return { tone: line.tone, text: t(line.key, { when: line.at ? whenText(line.at, now) : '' }) };
}
