/** Uptime of one service in percent; null without data (#81). */
export interface ServiceUptime {
  serviceId: string;
  day: number | null;
  week: number | null;
  month: number | null;
  quarter: number | null;
  /** Last 90 days, oldest first. */
  days: (number | null)[];
}

/** A confirmed outage. */
export interface Incident {
  id: string;
  serviceId: string;
  start: Date;
  end: Date | null;
  durationSeconds: number;
  cause: string;
}

/** "99.95 %" with as many decimals as useful; "–" without data. */
export function formatUptime(p: number | null): string {
  if (p === null) return '–';
  if (p >= 100) return '100 %';
  return `${p >= 99 ? p.toFixed(2) : p.toFixed(1)} %`;
}

/** Colour class of a day bar. */
export function uptimeLevel(p: number | null): 'none' | 'good' | 'degraded' | 'bad' {
  if (p === null) return 'none';
  if (p >= 99.5) return 'good';
  if (p >= 95) return 'degraded';
  return 'bad';
}

/** "3 h 20 min", "45 min" or "30 s". */
export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)} s`;
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes} min`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h < 48) return m ? `${h} h ${m} min` : `${h} h`;
  return `${Math.round(h / 24)} d`;
}
