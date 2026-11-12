/** Who wins when a discovered tile was edited by hand (#78). */
export type MergeRule = 'labels' | 'manual';

/** Outcome of the last discovery sync. */
export interface SyncResult {
  at: Date;
  /** Empty when the sync succeeded. */
  error: string;
  created: number;
  updated: number;
  hidden: number;
  shown: number;
  /** Containers with invalid labels, e.g. a URL without http(s). */
  skipped: number;
}

/** Editable discovery settings. */
export interface DiscoverySettings {
  enabled: boolean;
  intervalMinutes: number;
  rule: MergeRule;
}

/** State of Docker autodiscovery. */
export interface Discovery extends DiscoverySettings {
  /** False when the docker module is not enabled. */
  available: boolean;
  lastSync: SyncResult | null;
}

/** Interval limits in minutes, as enforced by the backend. */
export const MIN_INTERVAL = 1;
export const MAX_INTERVAL = 24 * 60;

/** Whether the interval is a whole number of minutes within the limits. */
export function validInterval(minutes: number): boolean {
  return Number.isInteger(minutes) && minutes >= MIN_INTERVAL && minutes <= MAX_INTERVAL;
}

/** Labels to copy into a compose file, shown as an example in the settings. */
export const labelExample = `labels:
  trisse-home.name: Jellyfin
  trisse-home.url: https://jellyfin.example.com
  trisse-home.icon: 🎬
  trisse-home.description: Films and series`;
