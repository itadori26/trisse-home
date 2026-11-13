import { t } from '../../../shared/i18n';
/** A subscribed feed. */
export interface Feed {
  id: string;
  url: string;
  title: string;
  siteUrl: string;
  lastError: string;
  lastFetchedAt: Date | null;
  unread: number;
}

/** A feed entry. */
export interface FeedItem {
  id: string;
  feedId: string;
  feedTitle: string;
  title: string;
  link: string;
  publishedAt: Date;
  read: boolean;
}

/** Configuration of the RSS widget. */
export interface RssConfig {
  feedId: string;
  unreadOnly: boolean;
  limit: number;
  /** Titles only: feed name and age move into the link's tooltip (#125). */
  compact: boolean;
}

/** Reads a stored widget configuration. */
export function parseRssConfig(raw: Record<string, unknown>): RssConfig {
  const limit = typeof raw['limit'] === 'number' ? Math.round(raw['limit']) : 20;
  return {
    feedId: typeof raw['feedId'] === 'string' ? raw['feedId'] : '',
    unreadOnly: raw['unreadOnly'] !== false,
    limit: Math.min(Math.max(limit, 5), 100),
    compact: raw['compact'] === true,
  };
}

/**
 * The "feed · age" line of an entry. The feed name is left out when the widget
 * shows a single feed, where it would only repeat itself.
 */
export function itemMeta(
  item: Pick<FeedItem, 'feedTitle' | 'publishedAt'>,
  singleFeed: boolean,
  now: Date,
): string {
  const when = age(item.publishedAt, now);
  return singleFeed || !item.feedTitle ? when : `${item.feedTitle} · ${when}`;
}

/** Formats an age such as "5 min", "3 h" or "2 d". */
export function age(date: Date, now: Date): string {
  const minutes = Math.max(0, Math.round((now.getTime() - date.getTime()) / 60_000));
  if (minutes < 60) return t('time.age_minutes', { n: minutes });
  const hours = Math.round(minutes / 60);
  if (hours < 48) return t('time.age_hours', { n: hours });
  return t('time.age_days', { n: Math.round(hours / 24) });
}
