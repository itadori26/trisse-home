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
  policy: NotificationPolicy;
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

/** Importance of a notification (#69). */
export type Severity = 'info' | 'warning' | 'critical';

/** Alert types that have a configurable severity. */
export const alertTypes = [
  'service',
  'container',
  'disk',
  'tls',
  'backup',
  'check',
  'security',
  'diskhealth',
  'forecast',
  'speedtest',
  'image',
] as const;

/** Severities per alert type and quiet hours. */
export interface NotificationPolicy {
  severities: Record<string, Severity>;
  /** Days are 0 (Sunday) to 6; empty means every day. */
  quiet: { enabled: boolean; start: string; end: string; days: number[] };
}

/** An additional delivery channel (#87); the secret is never returned. */
export interface Channel {
  id: string;
  kind: 'gotify' | 'webhook' | 'email';
  name: string;
  enabled: boolean;
  minSeverity: Severity;
  /** Empty: all alert types. */
  types: string[];
  url: string;
  hasSecret: boolean;
  template: string;
  smtpHost: string;
  smtpPort: number;
  username: string;
  from: string;
  to: string;
  tls: 'starttls' | 'tls' | 'none';
}

/** Input of a channel; an empty secret keeps the stored one. */
export type ChannelInput = Omit<Channel, 'id' | 'hasSecret'> & { secret: string };

/** A maintenance window that mutes notifications. */
export interface MaintenanceWindow {
  id: string;
  scope: string[];
  start: Date;
  end: Date;
  note: string;
}

/** Windows active at now. */
export function activeMaintenance(list: readonly MaintenanceWindow[], now: Date): MaintenanceWindow[] {
  return list.filter((w) => w.start <= now && now < w.end);
}

/** Turns "service:NAS, container:nginx" or "*" into scope entries. */
export function parseScope(input: string): string[] {
  return input
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
}

/** An empty channel of the given kind with sensible defaults. */
export function newChannel(kind: Channel['kind']): ChannelInput {
  return {
    kind,
    name: '',
    enabled: true,
    minSeverity: 'warning',
    types: [],
    url: '',
    secret: '',
    template: '',
    smtpHost: '',
    smtpPort: 587,
    username: '',
    from: '',
    to: '',
    tls: 'starttls',
  };
}
