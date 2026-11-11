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
