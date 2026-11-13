/** Media servers the widget can show (#61). */
export type MediaProvider = 'jellyfin' | 'plex';

/** Product name of a provider. */
export function providerName(p: MediaProvider): string {
  return p === 'plex' ? 'Plex' : 'Jellyfin';
}

/** Connection settings as shown in the UI (the token is never returned). */
export interface MediaSettings {
  configured: boolean;
  provider: MediaProvider;
  url: string;
  hasToken: boolean;
}

/** A recently added item; its poster is proxied by the backend. */
export interface MediaItem {
  id: string;
  title: string;
  series: string;
  kind: string;
  year: number;
}

/** Something that is being played right now. */
export interface PlaybackSession {
  user: string;
  device: string;
  title: string;
  series: string;
  kind: string;
  paused: boolean;
  /** 0–100 */
  progress: number;
}

/** What the widget shows. */
export interface MediaOverview {
  serverName: string;
  recent: MediaItem[];
  sessions: PlaybackSession[];
}

/** URL of the proxied poster of an item. */
export function posterUrl(id: string): string {
  return `/api/media/posters/${encodeURIComponent(id)}`;
}

/** "Show – Pilot" for episodes, otherwise the title. */
export function displayTitle(i: Pick<MediaItem, 'title' | 'series'>): string {
  return i.series ? `${i.series} – ${i.title}` : i.title;
}

/** Widget configuration. */
export interface MediaWidgetConfig {
  showRecent: boolean;
  showSessions: boolean;
}

/** Reads the widget configuration; both parts are shown by default. */
export function parseWidgetConfig(raw: Record<string, unknown>): MediaWidgetConfig {
  return { showRecent: raw['showRecent'] !== false, showSessions: raw['showSessions'] !== false };
}
