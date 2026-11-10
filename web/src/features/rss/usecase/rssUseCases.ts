import type { Feed, FeedItem, RssConfig } from '../domain/feed';

/** Output port: feed reader backend. */
export interface RssGateway {
  feeds(): Promise<Feed[]>;
  subscribe(url: string): Promise<Feed>;
  unsubscribe(id: string): Promise<void>;
  items(feedId: string, unreadOnly: boolean, limit: number): Promise<FeedItem[]>;
  setRead(itemId: string, read: boolean): Promise<void>;
  markAllRead(feedId: string): Promise<void>;
  refresh(): Promise<Feed[]>;
}

/** Application logic of the feed reader. */
export class RssUseCases {
  constructor(private readonly gateway: RssGateway) {}

  feeds(): Promise<Feed[]> {
    return this.gateway.feeds();
  }

  subscribe(url: string): Promise<Feed> {
    return this.gateway.subscribe(url.trim());
  }

  unsubscribe(id: string): Promise<void> {
    return this.gateway.unsubscribe(id);
  }

  items(cfg: RssConfig): Promise<FeedItem[]> {
    return this.gateway.items(cfg.feedId, cfg.unreadOnly, cfg.limit);
  }

  /** Marks an item as read when it is opened. */
  async open(item: FeedItem): Promise<FeedItem> {
    if (!item.read) await this.gateway.setRead(item.id, true);
    return { ...item, read: true };
  }

  toggleRead(item: FeedItem): Promise<void> {
    return this.gateway.setRead(item.id, !item.read);
  }

  markAllRead(feedId: string): Promise<void> {
    return this.gateway.markAllRead(feedId);
  }

  refresh(): Promise<Feed[]> {
    return this.gateway.refresh();
  }
}
