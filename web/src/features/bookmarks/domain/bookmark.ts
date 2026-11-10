/** A saved link. */
export interface Bookmark {
  id: string;
  folderId: string;
  title: string;
  url: string;
  description: string;
  tags: string[];
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
export type BookmarkInput = Omit<Bookmark, 'id'>;

/** Folder filter value that selects bookmarks without folder. */
export const UNFILED = 'none';

/** Configuration of the bookmarks widget. */
export interface BookmarksConfig {
  folderId: string;
  tag: string;
}

/** Reads a stored widget configuration. */
export function parseBookmarksConfig(raw: Record<string, unknown>): BookmarksConfig {
  return {
    folderId: typeof raw['folderId'] === 'string' ? raw['folderId'] : '',
    tag: typeof raw['tag'] === 'string' ? raw['tag'] : '',
  };
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
