import { t } from '../../../shared/i18n';
/** A saved link. */
export interface Bookmark {
  id: string;
  folderId: string;
  title: string;
  url: string;
  description: string;
  tags: string[];
  /** Result of the last link check (#72). */
  link: LinkStatus;
  /** A favicon is stored on the server. */
  hasIcon: boolean;
  /** ISO time the bookmark was added. */
  createdAt: string;
}

/** Result of a link check; state '' = not checked yet. */
export interface LinkStatus {
  state: '' | 'ok' | 'redirect' | 'broken';
  code: number;
  /** Final URL of a redirect. */
  target: string;
  /** ISO time of the check ('' = never). */
  checkedAt: string;
}

/** Outbound features of the bookmarks module; both off by default (#72). */
export interface BookmarkSettings {
  enrich: boolean;
  checkLinks: boolean;
}

/**
 * Icon source of a bookmark: the copy stored on the server, otherwise the
 * site's own /favicon.ico (loaded by the browser).
 */
export function iconSrc(b: Pick<Bookmark, 'id' | 'url' | 'hasIcon'>): string | null {
  if (b.hasIcon) return `/api/bookmarks/${encodeURIComponent(b.id)}/icon`;
  return faviconUrl(b.url);
}

/** Describes a link problem for a tooltip, or '' when the link is fine. */
export function linkProblem(link: LinkStatus, locale?: string): string {
  const when = link.checkedAt ? new Date(link.checkedAt).toLocaleDateString(locale) : '';
  if (link.state === 'broken')
    return link.code
      ? t('bookmarks.link_broken_code', { code: link.code, when })
      : t('bookmarks.link_unreachable', { when });
  if (link.state === 'redirect') return t('bookmarks.link_redirects', { target: link.target, when });
  return '';
}

/** A folder with its bookmark count. */
export interface Folder {
  id: string;
  name: string;
  count: number;
}

/** A tag with its usage count. */
export interface Tag {
  name: string;
  count: number;
}

/** Editable fields of a bookmark. */
export type BookmarkInput = Pick<Bookmark, 'folderId' | 'title' | 'url' | 'description' | 'tags'>;

/** Folder filter value that selects bookmarks without folder. */
export const UNFILED = 'none';

/** Configuration of the bookmarks widget. */
export interface BookmarksConfig {
  folderId: string;
  tag: string;
  sort: BookmarkSort;
}

/** Sort order of the widget: as delivered by the server, A-Z, or newest first (#120). */
export type BookmarkSort = 'manual' | 'az' | 'newest';

const SORTS: readonly BookmarkSort[] = ['manual', 'az', 'newest'];

/** Reads a stored widget configuration. */
export function parseBookmarksConfig(raw: Record<string, unknown>): BookmarksConfig {
  return {
    folderId: typeof raw['folderId'] === 'string' ? raw['folderId'] : '',
    tag: typeof raw['tag'] === 'string' ? raw['tag'] : '',
    sort: SORTS.find((s) => s === raw['sort']) ?? 'manual',
  };
}

/**
 * Orders bookmarks for display without changing the input. A-Z compares
 * titles with the locale's rules (case-insensitive, numbers by value, so
 * "Server 2" comes before "Server 10"); newest first uses the date added and
 * falls back to A-Z for equal dates.
 */
export function sortBookmarks<T extends Pick<Bookmark, 'title' | 'createdAt'>>(
  list: readonly T[],
  sort: BookmarkSort,
  locale?: string,
): T[] {
  if (sort === 'manual') return [...list];
  const collator = new Intl.Collator(locale, { sensitivity: 'base', numeric: true });
  const byTitle = (a: T, b: T) => collator.compare(a.title, b.title);
  if (sort === 'az') return [...list].sort(byTitle);
  const time = (b: T) => Date.parse(b.createdAt) || 0;
  return [...list].sort((a, b) => time(b) - time(a) || byTitle(a, b));
}

/** Splits user input like "go, web #server" into tags. */
export function parseTags(input: string): string[] {
  const tags = input
    .split(/[\s,]+/)
    .map((t) => t.replace(/^#/, '').trim().toLowerCase())
    .filter((t) => t !== '');
  return [...new Set(tags)];
}

/** Returns the favicon URL of a bookmark's host via the site itself. */
export function faviconUrl(url: string): string | null {
  try {
    const u = new URL(url);
    return `${u.protocol}//${u.host}/favicon.ico`;
  } catch {
    return null;
  }
}
