/** DNS filters the widget can show (#83). */
export type DnsProvider = 'adguard' | 'pihole';

/** Every provider, in the order of the settings. */
export const dnsProviders: readonly DnsProvider[] = ['adguard', 'pihole'];

/** Product name of a provider. */
export function providerName(p: DnsProvider): string {
  return p === 'pihole' ? 'Pi-hole' : 'AdGuard Home';
}

/** Whether the provider needs a username (Pi-hole uses a password only). */
export function needsUsername(p: DnsProvider): boolean {
  return p === 'adguard';
}

/** Connection settings as shown in the UI (the password is never returned). */
export interface AdGuardSettings {
  provider: DnsProvider;
  configured: boolean;
  url: string;
  username: string;
  hasPassword: boolean;
  /** The server allows pausing this provider (DNS_PAUSE_PROVIDERS, #177). */
  canPause: boolean;
}

/** DNS filter statistics, the same for every provider. */
export interface AdGuardStats {
  queries: number;
  blocked: number;
  blockedPercent: number;
  /** Null when the provider does not report it (Pi-hole). */
  avgProcessingMs: number | null;
  protectionEnabled: boolean;
  /** End of a timed pause (#177), null otherwise. */
  pausedUntil: Date | null;
  topBlocked: { domain: string; count: number }[];
  error: string;
}

/** Widget configuration. */
export interface AdGuardWidgetConfig {
  provider: DnsProvider;
}

/** Reads the widget configuration; widgets from before #83 show AdGuard Home. */
export function parseWidgetConfig(raw: Record<string, unknown>): AdGuardWidgetConfig {
  return { provider: raw['provider'] === 'pihole' ? 'pihole' : 'adguard' };
}

/** Offered pause durations in minutes (#177). */
export const PAUSE_MINUTES = [5, 30, 60] as const;
export type PauseMinutes = (typeof PAUSE_MINUTES)[number];

/** Whole minutes left of a pause (at least 1 while it runs), 0 when over. */
export function pauseMinutesLeft(until: Date | null, now: Date): number {
  if (!until) return 0;
  const ms = until.getTime() - now.getTime();
  return ms <= 0 ? 0 : Math.max(1, Math.ceil(ms / 60_000));
}
