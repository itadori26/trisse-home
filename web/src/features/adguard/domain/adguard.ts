/** Connection settings as shown in the UI (the password is never returned). */
export interface AdGuardSettings {
  configured: boolean;
  url: string;
  username: string;
  hasPassword: boolean;
}

/** AdGuard Home statistics. */
export interface AdGuardStats {
  queries: number;
  blocked: number;
  blockedPercent: number;
  avgProcessingMs: number;
  protectionEnabled: boolean;
  topBlocked: { domain: string; count: number }[];
  error: string;
}
