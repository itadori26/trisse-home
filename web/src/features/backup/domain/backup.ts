/** A stored database backup. */
export interface BackupFile {
  name: string;
  size: number;
  createdAt: Date;
}

/** URL to download a backup (served with Content-Disposition: attachment). */
export function downloadUrl(file: BackupFile): string {
  return `/api/backups/${encodeURIComponent(file.name)}`;
}

/** Off-site WebDAV target (#46); the password is never sent back. */
export interface RemoteSettings {
  configured: boolean;
  url: string;
  username: string;
  hasPassword: boolean;
  retention: number;
}

/** Input of the off-site settings; an empty password keeps the stored one. */
export interface RemoteSettingsInput {
  url: string;
  username: string;
  password: string;
  retention: number;
}

/** Where a restored backup comes from. */
export type RestoreSource = 'local' | 'remote' | 'upload';

/** A restore request; password (and code with 2FA) are re-entered. */
export interface RestoreRequest {
  password: string;
  code: string;
  source: RestoreSource;
  name?: string;
  file?: File;
}

/** Builds the multipart body; the file must come last (it is streamed). */
export function restoreForm(req: RestoreRequest): FormData {
  const form = new FormData();
  form.append('password', req.password);
  form.append('code', req.code.trim());
  form.append('source', req.source);
  if (req.name) form.append('name', req.name);
  if (req.file) form.append('file', req.file);
  return form;
}

/** Traffic light of the backups (#677), judged by the server. */
export type BackupHealth = 'ok' | 'stale' | 'failed';

/** GET /api/backups/status; null times mean "never" or "not since the start". */
export interface BackupStatus {
  health: BackupHealth;
  lastBackupAt: Date | null;
  count: number;
  failedAt: Date | null;
  error: string;
  offsiteFailedAt: Date | null;
  offsiteError: string;
}

/** Colour of the traffic light: green, amber or red. */
export type HealthTone = 'ok' | 'warn' | 'error';

/** The one line the traffic light says, as a message key with parameters. */
export interface HealthLine {
  tone: HealthTone;
  key:
    | 'backup.health.ok'
    | 'backup.health.stale'
    | 'backup.health.none'
    | 'backup.health.offsite_failed'
    | 'backup.health.failed';
  /** The moment the line talks about; formatted by the caller. */
  at: Date | null;
}

/**
 * Says what the traffic light means in one line: a failed attempt first,
 * then a missing backup, a failed off-site copy, an overdue backup, and
 * otherwise "all good" with the time of the newest backup.
 */
export function healthLine(s: BackupStatus): HealthLine {
  if (s.health === 'failed') return { tone: 'error', key: 'backup.health.failed', at: s.failedAt };
  if (!s.lastBackupAt) return { tone: 'warn', key: 'backup.health.none', at: null };
  if (s.health === 'stale' && s.offsiteFailedAt)
    return { tone: 'warn', key: 'backup.health.offsite_failed', at: s.lastBackupAt };
  if (s.health === 'stale') return { tone: 'warn', key: 'backup.health.stale', at: s.lastBackupAt };
  return { tone: 'ok', key: 'backup.health.ok', at: s.lastBackupAt };
}

/** Which day a moment falls on, relative to now: for "today 03:00" and the like. */
export type DayWord = 'today' | 'yesterday' | 'earlier';

export function dayWord(at: Date, now: Date): DayWord {
  if (at.toDateString() === now.toDateString()) return 'today';
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  return at.toDateString() === yesterday.toDateString() ? 'yesterday' : 'earlier';
}
