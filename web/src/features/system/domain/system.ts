/** Used/total pair in bytes. */
export interface Usage {
  total: number;
  used: number;
  percent: number;
}

/** Host metrics snapshot. */
export interface SystemSnapshot {
  cpuPercent: number;
  cpuCount: number;
  load: [number, number, number];
  memory: Usage;
  swap: Usage;
  disks: { label: string; usage: Usage; forecast: DiskForecast | null }[];
  uptimeSeconds: number;
  temperature: number;
}

/** When a file system is full (#108); null before the first aggregate. */
export interface DiskForecast {
  status: 'filling' | 'flat' | 'insufficient';
  daysUntilFull: number;
  bytesPerDay: number;
  confidence: number;
}

/** Regression window and warning threshold of the disk forecast. */
export interface ForecastSettings {
  windowDays: number;
  alertDays: number;
}

/** Whole days until full, or null when there is no forecast to show. */
export function daysUntilFull(f: DiskForecast | null): number | null {
  if (!f || f.status !== 'filling') return null;
  return Math.max(0, Math.round(f.daysUntilFull));
}

/** Normalised load: 1-minute load per CPU in percent (100 = fully busy). */
export function loadPercent(s: SystemSnapshot): number {
  if (s.cpuCount <= 0) return 0;
  return Math.min(100, (s.load[0] / s.cpuCount) * 100);
}

/** One point of the 24 h metric history (percentages). */
export interface HistoryPoint {
  at: Date;
  cpu: number;
  memory: number;
  disk: number;
}

/** Running version and the result of the optional update check. */
export interface VersionInfo {
  current: string;
  checkEnabled: boolean;
  latest: { version: string; url: string; publishedAt: Date } | null;
  updateAvailable: boolean;
}
