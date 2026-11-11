/** Connection settings as shown in the UI (the app password is never returned). */
export interface NextcloudSettings {
  configured: boolean;
  url: string;
  username: string;
  hasPassword: boolean;
}

/** Server details from the serverinfo app (admins only). */
export interface NextcloudServer {
  freeSpace: number;
  users: number;
  activeUsers5m: number;
  activeUsers1d: number;
  files: number;
  appUpdates: number;
  databaseSize: number;
}

/** Status of the Nextcloud instance. */
export interface NextcloudStatus {
  version: string;
  maintenance: boolean;
  quota: { used: number; free: number; total: number; percent: number };
  server: NextcloudServer | null;
  serverHint: string;
  error: string;
}

/** Major.minor.patch without the build suffix, e.g. "30.0.2". */
export function shortVersion(v: string): string {
  return v.split(/[\s+-]/)[0] ?? v;
}
