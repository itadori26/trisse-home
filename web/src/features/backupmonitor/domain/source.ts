import { t } from '../../../shared/i18n';
/** State of a monitored backup. */
export type SourceState = 'ok' | 'overdue' | 'failed' | 'waiting' | 'running';

/** A backup job that reports via its ping URL. */
export interface BackupSource {
  id: string;
  name: string;
  /** Interval schedule; unused when cron is set. */
  maxAgeHours: number;
  kind: 'backup' | 'check';
  cron: string;
  graceMinutes: number;
  lastStartAt: Date | null;
  lastDurationSeconds: number;
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

/** When a source must report (#100). */
export interface SourceSchedule {
  kind: 'backup' | 'check';
  /** Used when cron is empty. */
  maxAgeHours: number;
  /** Five-field cron expression such as "0 3 * * *". */
  cron: string;
  graceMinutes: number;
}

/** Example commands for a script; /start measures the duration. */
export function pingExamples(pingUrl: string): string {
  return [
    `# optional, before the job starts (measures the duration):`,
    `curl -fsS -X POST ${pingUrl}/start`,
    `# after a successful run:`,
    `curl -fsS -X POST ${pingUrl}`,
    `# after a failed run (optional message in the body):`,
    `curl -fsS -X POST --data "disk full" ${pingUrl}/fail`,
  ].join('\n');
}

/** Rough client-side check of a cron expression (the server validates fully). */
export function looksLikeCron(expr: string): boolean {
  const e = expr.trim();
  return /^@(hourly|daily|midnight|weekly|monthly)$/i.test(e) || /^([\d*,/-]+\s+){4}[\d*,/-]+$/.test(e);
}

/** "3 h ago", "2 d ago" or "never". */
export function ago(date: Date | null, now: Date): string {
  if (!date) return t('time.never');
  const hours = Math.max(0, (now.getTime() - date.getTime()) / 3_600_000);
  if (hours < 1) return t('time.minutes_ago', { n: Math.round(hours * 60) });
  if (hours < 48) return t('time.hours_ago', { n: Math.round(hours) });
  return t('time.days_ago', { n: Math.round(hours / 24) });
}
