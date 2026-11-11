/** A public transport stop. */
export interface Stop {
  id: string;
  name: string;
  lat: number;
  lon: number;
}

/** One departure. */
export interface Departure {
  line: string;
  headsign: string;
  mode: string;
  planned: Date;
  expected: Date | null;
  delayMinutes: number;
  track: string;
  cancelled: boolean;
}

/** Configuration of a transit widget. */
export interface TransitConfig {
  stopId: string;
  stopName: string;
  count: number;
}

/** Reads a stored widget configuration (null without a stop). */
export function parseTransitConfig(raw: Record<string, unknown>): TransitConfig | null {
  if (typeof raw['stopId'] !== 'string' || raw['stopId'] === '') return null;
  const count = typeof raw['count'] === 'number' ? Math.round(raw['count']) : 6;
  return {
    stopId: raw['stopId'],
    stopName: typeof raw['stopName'] === 'string' ? raw['stopName'] : '',
    count: Math.min(Math.max(count, 1), 20),
  };
}

/** Minutes until departure ("now" below one minute). */
export function minutesUntil(d: Pick<Departure, 'planned' | 'expected'>, now: Date): string {
  const at = d.expected ?? d.planned;
  const min = Math.round((at.getTime() - now.getTime()) / 60_000);
  return min <= 0 ? 'now' : `${min} min`;
}

/** Icon per transport mode (MOTIS modes). */
export function modeIcon(mode: string): string {
  if (mode.includes('bus')) return '🚌';
  if (mode.includes('tram')) return '🚊';
  if (mode.includes('subway') || mode.includes('metro')) return '🚇';
  if (mode.includes('ferry')) return '⛴️';
  if (
    mode.includes('rail') ||
    mode.includes('train') ||
    mode.includes('regional') ||
    mode.includes('suburban')
  )
    return '🚆';
  return '🚏';
}
