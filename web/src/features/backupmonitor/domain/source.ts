import { t } from '../../../shared/i18n';
/** State of a monitored backup. */
export type SourceState = 'ok' | 'overdue' | 'failed' | 'waiting';

/** A backup job that reports via its ping URL. */
export interface BackupSource {
  id: string;
  name: string;
  maxAgeHours: number;
  state: SourceState;
  lastSuccessAt: Date | null;
  lastFailureAt: Date | null;
  lastMessage: string;
}

/** Shown once after creating a source or rotating its token. */
export interface CreatedSource {
  source: BackupSource;
  pingUrl: string;
}

/** Example commands for a backup script. */
export function pingExamples(pingUrl: string): string {
  return [
    `# after a successful run:`,
    `curl -fsS -X POST ${pingUrl}`,
    `# after a failed run (optional message in the body):`,
    `curl -fsS -X POST --data "disk full" "${pingUrl}?status=fail"`,
  ].join('\n');
}

/** "3 h ago", "2 d ago" or "never". */
export function ago(date: Date | null, now: Date): string {
  if (!date) return t('time.never');
  const hours = Math.max(0, (now.getTime() - date.getTime()) / 3_600_000);
  if (hours < 1) return t('time.minutes_ago', { n: Math.round(hours * 60) });
  if (hours < 48) return t('time.hours_ago', { n: Math.round(hours) });
  return t('time.days_ago', { n: Math.round(hours / 24) });
}
