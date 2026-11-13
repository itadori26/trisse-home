/**
 * Launcher (#189): one overview of services and bookmarks with a filter,
 * segments and short sections. Pure rules; the data comes through ports.
 */

export type LauncherContent = 'both' | 'services' | 'bookmarks';
export type Segment = 'all' | 'services' | 'bookmarks';
export type ServiceState = 'online' | 'offline' | 'unknown';

export interface LauncherService {
  id: string;
  name: string;
  url: string;
  state: ServiceState;
  latencyMs: number;
}

export interface LauncherBookmark {
  id: string;
  title: string;
  url: string;
  folder: string;
}

/** Rows shown per section before "Show all". */
export const SERVICE_PREVIEW = 6;
export const BOOKMARK_PREVIEW = 6;

/** Lower-cased text without diacritics. */
const fold = (s: string) =>
  s
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase();

/** Whether every word of the filter occurs in one of the texts. */
export function matches(filter: string, ...texts: string[]): boolean {
  const words = fold(filter).split(/\s+/).filter(Boolean);
  const hay = fold(texts.join(' '));
  return words.every((w) => hay.includes(w));
}

/** Only http(s) links are opened from the launcher. */
export function safeUrl(url: string): string | null {
  try {
    const u = new URL(url);
    return u.protocol === 'http:' || u.protocol === 'https:' ? u.href : null;
  } catch {
    return null;
  }
}

export interface Section<T> {
  /** Items to show (all, or the first rows). */
  items: T[];
  /** Number of matching items. */
  total: number;
  /** Whether "Show all" is offered (items are hidden). */
  more: boolean;
}

function section<T>(list: T[], limit: number, expanded: boolean): Section<T> {
  return {
    items: expanded ? list : list.slice(0, limit),
    total: list.length,
    more: !expanded && list.length > limit,
  };
}

export interface LauncherView {
  services: Section<LauncherService> | null;
  bookmarks: Section<LauncherBookmark> | null;
  /** Counts for the segment buttons. */
  counts: Record<Segment, number>;
}

/**
 * Builds the launcher for a filter and segment. The content setting hides
 * a kind completely; "Show all" expands one section.
 */
export function buildView(
  services: readonly LauncherService[],
  bookmarks: readonly LauncherBookmark[],
  opts: {
    filter: string;
    segment: Segment;
    content: LauncherContent;
    expanded: Set<'services' | 'bookmarks'>;
  },
): LauncherView {
  const withServices = opts.content !== 'bookmarks';
  const withBookmarks = opts.content !== 'services';
  const s = withServices ? services.filter((x) => matches(opts.filter, x.name, x.url)) : [];
  const b = withBookmarks ? bookmarks.filter((x) => matches(opts.filter, x.title, x.url, x.folder)) : [];
  const counts = { all: s.length + b.length, services: s.length, bookmarks: b.length };
  // A segment alone shows its full list.
  const only = (kind: 'services' | 'bookmarks') => opts.segment === kind || opts.expanded.has(kind);
  return {
    services:
      withServices && opts.segment !== 'bookmarks' ? section(s, SERVICE_PREVIEW, only('services')) : null,
    bookmarks:
      withBookmarks && opts.segment !== 'services' ? section(b, BOOKMARK_PREVIEW, only('bookmarks')) : null,
    counts,
  };
}

/** Summary of the status pill: all online, or how many are offline. */
export function statusSummary(services: readonly Pick<LauncherService, 'state'>[]): {
  kind: 'none' | 'ok' | 'offline';
  offline: number;
} {
  if (services.length === 0) return { kind: 'none', offline: 0 };
  const offline = services.filter((s) => s.state === 'offline').length;
  return { kind: offline > 0 ? 'offline' : 'ok', offline };
}

/** The first letter (or digit) of a name for the avatar of a bookmark. */
export function initial(name: string): string {
  const m = name.trim().match(/[\p{L}\p{N}]/u);
  return (m?.[0] ?? '?').toUpperCase();
}
