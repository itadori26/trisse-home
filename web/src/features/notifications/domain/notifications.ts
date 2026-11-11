/** What is notified. */
export interface NotificationRules {
  services: boolean;
  containers: boolean;
  disk: boolean;
  diskThreshold: number;
  tls: boolean;
  tlsDays: number;
  backups: boolean;
  security: boolean;
}

/** Notification settings as shown in the UI (the token is never returned). */
export interface NotificationSettings {
  enabled: boolean;
  serverUrl: string;
  topic: string;
  hasToken: boolean;
  rules: NotificationRules;
  lastError: string;
  lastDelivered: Date | null;
}

/** Suggests a hard-to-guess topic name; topics on public ntfy servers are public. */
export function suggestTopic(random: () => number = Math.random): string {
  const chars = 'abcdefghijkmnpqrstuvwxyz23456789';
  let s = 'trisse-home-';
  for (let i = 0; i < 12; i++) s += chars[Math.floor(random() * chars.length)];
  return s;
}
