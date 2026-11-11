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

/** Status of the box. */
export interface FritzBoxStatus {
  model: string;
  wan: Wan;
  calls: Call[];
  hint: string;
  error: string;
}

/** Formats bytes per second as Mbit/s (or kbit/s below 1 Mbit/s). */
export function formatRate(bytesPerSecond: number): string {
  const bits = bytesPerSecond * 8;
  if (bits < 1_000_000) return `${Math.round(bits / 1000)} kbit/s`;
  return `${(bits / 1_000_000).toFixed(bits < 10_000_000 ? 1 : 0)} Mbit/s`;
}

/** Display name of the other party. */
export function callParty(c: Pick<Call, 'name' | 'number'>): string {
  return c.name || c.number || 'Unknown';
}

/** Arrow/icon per call type. */
export const callIcons: Record<CallType, string> = {
  incoming: '↙',
  missed: '✗',
  outgoing: '↗',
  rejected: '⊘',
  active: '☎',
};
