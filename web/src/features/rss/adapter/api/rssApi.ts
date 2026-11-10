import type { HttpClient } from '../../../../shared/api-client/http';
import type { FeedDTO, ItemDTO } from '../../../../shared/api-client/generated/rss';
import type {
  MarkAllReadDTO,
  ReadDTO,
  SubscribeDTO,
} from '../../../../shared/api-client/generated/rss-requests';
import type { Feed, FeedItem } from '../../domain/feed';
import type { RssGateway } from '../../usecase/rssUseCases';

const toFeed = (d: FeedDTO): Feed => ({
  id: d.id,
  url: d.url,
  title: d.title,
  siteUrl: d.siteUrl,
  lastError: d.lastError,
  lastFetchedAt: d.lastFetchedAt ? new Date(d.lastFetchedAt) : null,
  unread: d.unread,
});

const toItem = (d: ItemDTO): FeedItem => ({
  id: d.id,
  feedId: d.feedId,
  feedTitle: d.feedTitle,
  title: d.title,
  link: d.link,
  publishedAt: new Date(d.publishedAt),
  read: d.read,
});

/** RssGateway backed by the REST API. */
export class RssApi implements RssGateway {
  constructor(private readonly http: HttpClient) {}

  async feeds(): Promise<Feed[]> {
    return (await this.http.get<FeedDTO[]>('/api/rss/feeds')).map(toFeed);
  }
  async subscribe(url: string): Promise<Feed> {
    const body: SubscribeDTO = { url };
    return toFeed(await this.http.post<FeedDTO>('/api/rss/feeds', body));
  }
  unsubscribe(id: string): Promise<void> {
    return this.http.delete(`/api/rss/feeds/${encodeURIComponent(id)}`);
  }
  async items(feedId: string, unreadOnly: boolean, limit: number): Promise<FeedItem[]> {
    const q = new URLSearchParams({ limit: String(limit) });
    if (feedId) q.set('feed', feedId);
    if (unreadOnly) q.set('unread', 'true');
    return (await this.http.get<ItemDTO[]>(`/api/rss/items?${q}`)).map(toItem);
  }
  setRead(itemId: string, read: boolean): Promise<void> {
    const body: ReadDTO = { read };
    return this.http.put(`/api/rss/items/${encodeURIComponent(itemId)}/read`, body);
  }
  markAllRead(feedId: string): Promise<void> {
    const body: MarkAllReadDTO = { feedId };
    return this.http.post('/api/rss/read-all', body);
  }
  async refresh(): Promise<Feed[]> {
    return (await this.http.post<FeedDTO[]>('/api/rss/refresh')).map(toFeed);
  }
}
