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
