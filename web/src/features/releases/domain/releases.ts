/** A MusicBrainz artist. */
export interface Artist {
  id: string;
  name: string;
  /** Tells artists with the same name apart, e.g. "UK band". */
  disambiguation: string;
}

/** A new album, single or EP; its cover is proxied by the backend. */
export interface Release {
  id: string;
  title: string;
  artist: string;
  type: string;
  /** YYYY-MM-DD */
  date: string;
}

/** The release list with status information. */
export interface ReleaseList {
  releases: Release[];
  error: string;
}

/** URL of the proxied cover of a release. */
export function coverUrl(id: string): string {
  return `/api/releases/covers/${encodeURIComponent(id)}`;
}

/** Widget configuration. */
export interface ReleasesWidgetConfig {
  weeks: number;
}

/** Reads the widget configuration (1–12 weeks, default 4). */
export function parseWidgetConfig(raw: Record<string, unknown>): ReleasesWidgetConfig {
  const w = typeof raw['weeks'] === 'number' ? Math.round(raw['weeks']) : 4;
  return { weeks: Math.min(Math.max(w, 1), 12) };
}

/** "Radiohead (UK band)" for the search results. */
export function artistLabel(a: Pick<Artist, 'name' | 'disambiguation'>): string {
  return a.disambiguation ? `${a.name} (${a.disambiguation})` : a.name;
}
