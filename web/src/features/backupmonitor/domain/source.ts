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
  if (!date) return 'never';
  const hours = Math.max(0, (now.getTime() - date.getTime()) / 3_600_000);
  if (hours < 1) return `${Math.round(hours * 60)} min ago`;
  if (hours < 48) return `${Math.round(hours)} h ago`;
  return `${Math.round(hours / 24)} d ago`;
}
