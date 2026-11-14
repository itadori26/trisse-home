import { t } from '../../../shared/i18n';
/** Connection settings as shown in the UI (the password is never returned). */
export interface FritzBoxSettings {
  configured: boolean;
  url: string;
  username: string;
  hasPassword: boolean;
}

/** Kind of a call. */
export type CallType = 'incoming' | 'missed' | 'outgoing' | 'rejected' | 'active';

/** One entry of the call list. */
export interface Call {
  type: CallType;
  number: string;
  name: string;
  at: Date;
  durationMinutes: number;
}

/** Internet connection; rates in bytes/s, link speeds in bits/s. */
export interface Wan {
  connected: boolean;
  externalIp: string;
  uptimeSeconds: number;
  throughputKnown: boolean;
  downstreamBytesPerSecond: number;
  upstreamBytesPerSecond: number;
  downstreamMaxBits: number;
  upstreamMaxBits: number;
  downstreamPercent: number;
  upstreamPercent: number;
}

/** Status of the internet connection (the calls are separate, #168). */
export interface FritzBoxStatus {
  model: string;
  wan: Wan;
  hint: string;
  error: string;
}

/** Which calls a calls widget shows (#168). */
export type CallFilterKind = 'all' | 'missed' | 'incoming' | 'outgoing';
export const CALL_FILTERS: readonly CallFilterKind[] = ['all', 'missed', 'incoming', 'outgoing'];

/** The call list of a calls widget. */
export interface CallList {
  calls: Call[];
  /** Explains an empty list, e.g. no login configured. */
  hint: string;
  /** Set when the box did not answer; the calls are the last known ones. */
  error: string;
  updatedAt: Date | null;
}

/** What the internet widget shows (#168); everything unless switched off. */
export interface InternetOptions {
  showIp: boolean;
  showRates: boolean;
  showUptime: boolean;
}

export function internetOptions(config: Record<string, unknown>): InternetOptions {
  return {
    showIp: config['showIp'] !== false,
    showRates: config['showRates'] !== false,
    showUptime: config['showUptime'] !== false,
  };
}

/** What the calls widget shows (#168). */
export interface CallsOptions {
  filter: CallFilterKind;
  /** 5 to 50, default 10. */
  limit: number;
  /** Numbers of unknown callers; off shows only names. */
  showNumbers: boolean;
}

export const DEFAULT_CALLS = 10;

export function callsOptions(config: Record<string, unknown>): CallsOptions {
  const f = config['filter'];
  const n = Number(config['limit']);
  return {
    filter: CALL_FILTERS.includes(f as CallFilterKind) ? (f as CallFilterKind) : 'all',
    limit: Number.isInteger(n) && n >= 5 && n <= 50 ? n : DEFAULT_CALLS,
    showNumbers: config['showNumbers'] !== false,
  };
}

/** Formats bytes per second as Mbit/s (or kbit/s below 1 Mbit/s). */
export function formatRate(bytesPerSecond: number): string {
  const bits = bytesPerSecond * 8;
  if (bits < 1_000_000) return `${Math.round(bits / 1000)} kbit/s`;
  return `${(bits / 1_000_000).toFixed(bits < 10_000_000 ? 1 : 0)} Mbit/s`;
}

/** Display name of the other party; numbers can be hidden (#168). */
export function callParty(c: Pick<Call, 'name' | 'number'>, showNumbers = true): string {
  return c.name || (showNumbers && c.number) || t('fritzbox.unknown_caller');
}

/** Arrow/icon per call type. */
export const callIcons: Record<CallType, string> = {
  incoming: '↙',
  missed: '✗',
  outgoing: '↗',
  rejected: '⊘',
  active: '☎',
};
