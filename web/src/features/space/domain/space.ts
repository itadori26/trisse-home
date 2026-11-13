import { t } from '../../../shared/i18n';

export type LaunchStatus =
  'go' | 'tbd' | 'tbc' | 'hold' | 'in_flight' | 'success' | 'failure' | 'partial_failure' | 'unknown';

/** An upcoming launch (#64). */
export interface Launch {
  id: string;
  mission: string;
  missionType: string;
  description: string;
  rocket: string;
  provider: string;
  pad: string;
  location: string;
  net: Date;
  status: LaunchStatus;
  webcastLive: boolean;
}

/** The picture of the day; image is the proxied URL ('' on video days). */
export interface Picture {
  date: string;
  title: string;
  explanation: string;
  copyright: string;
  mediaType: string;
  image: string;
  link: string;
}

export interface SpaceSettings {
  launches: boolean;
  picture: boolean;
  hasKey: boolean;
}

export type SpaceShow = 'both' | 'launches' | 'picture';

/** Configuration of the widget. */
export interface SpaceConfig {
  show: SpaceShow;
  count: number;
}

export function parseSpaceConfig(raw: Record<string, unknown>): SpaceConfig {
  const show = raw['show'] === 'launches' || raw['show'] === 'picture' ? raw['show'] : 'both';
  const count = typeof raw['count'] === 'number' ? Math.round(raw['count']) : 4;
  return { show, count: Math.min(Math.max(count, 1), 10) };
}

/** "in 3 h", "in 2 days", "now": time until a launch. */
export function countdown(net: Date, now: Date): string {
  const minutes = Math.round((net.getTime() - now.getTime()) / 60_000);
  if (minutes <= 0) return t('space.now');
  if (minutes < 60) return t('space.in_minutes', { n: minutes });
  const hours = Math.round(minutes / 60);
  if (hours < 48) return t('space.in_hours', { n: hours });
  return t('space.in_days', { n: Math.round(hours / 24) });
}

/** Statuses shown as confirmed (green) or uncertain. */
export const statusTone = (s: LaunchStatus): 'good' | 'bad' | 'neutral' =>
  s === 'go' || s === 'success'
    ? 'good'
    : s === 'hold' || s === 'failure' || s === 'partial_failure'
      ? 'bad'
      : 'neutral';

/** The stable background address of today's picture on this server. */
export const backgroundUrl = (origin: string): string => `${origin}/api/space/apod/image`;
