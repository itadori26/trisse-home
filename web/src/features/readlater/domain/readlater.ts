/** A saved article (#99). */
export interface Article {
  id: string;
  url: string;
  title: string;
  site: string;
  excerpt: string;
  /** Sanitised reader copy (HTML); only set for a single article. */
  content: string;
  readingMinutes: number;
  status: ArticleStatus;
  savedAt: Date;
  bytes: number;
  /** Why there is no reader copy (the link is kept). */
  error: string;
}

export type ArticleStatus = 'unread' | 'read' | 'archived';
export const STATUSES: readonly ArticleStatus[] = ['unread', 'read', 'archived'];

/** Whether a text looks like an http(s) address worth sending. */
export const looksLikeUrl = (s: string): boolean => /^https?:\/\/\S+\.\S+/i.test(s.trim());

/** Configuration of the widget. */
export function parseReadLaterConfig(raw: Record<string, unknown>): { count: number } {
  const count = typeof raw['count'] === 'number' ? Math.round(raw['count']) : 6;
  return { count: Math.min(Math.max(count, 1), 30) };
}

/** "1.2 MB" style sizes for the storage overview. */
export function formatBytes(n: number, locale: string): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toLocaleString(locale, { maximumFractionDigits: 0 })} KB`;
  return `${(n / 1024 / 1024).toLocaleString(locale, { maximumFractionDigits: 1 })} MB`;
}
