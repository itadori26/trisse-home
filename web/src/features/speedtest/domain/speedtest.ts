/** One Speedtest Tracker result (#60). */
export interface Measurement {
  at: Date;
  downloadMbps: number;
  uploadMbps: number;
  pingMs: number;
}

/** Latest result and history, oldest first. */
export interface SpeedResults {
  latest: Measurement | null;
  history: Measurement[];
  error: string;
}

/** Connection settings; the token is never sent back. */
export interface SpeedtestSettings {
  configured: boolean;
  url: string;
  hasToken: boolean;
  minDownMbps: number;
  minUpMbps: number;
}

/** History ranges of the widget in hours. */
export const ranges = [24, 168] as const;
export type Range = (typeof ranges)[number];

/** Formats Mbit/s with sensible precision. */
export function formatMbps(v: number): string {
  return v >= 100 ? v.toFixed(0) : v.toFixed(1);
}

/** Whether the latest result breaks a configured threshold. */
export function belowThreshold(
  m: Measurement | null,
  s: Pick<SpeedtestSettings, 'minDownMbps' | 'minUpMbps'>,
): boolean {
  if (!m) return false;
  return (
    (s.minDownMbps > 0 && m.downloadMbps < s.minDownMbps) || (s.minUpMbps > 0 && m.uploadMbps < s.minUpMbps)
  );
}
