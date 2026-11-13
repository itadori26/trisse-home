import { t } from '../../../shared/i18n';
import { type Feed, type FeedItem, type ImportReport, MAX_OPML_BYTES, type RssConfig } from '../domain/feed';

/** Output port: feed reader backend. */
export interface RssGateway {
  feeds(): Promise<Feed[]>;
  subscribe(url: string): Promise<Feed>;
  unsubscribe(id: string): Promise<void>;
  items(feedId: string, unreadOnly: boolean, limit: number): Promise<FeedItem[]>;
  setRead(itemId: string, read: boolean): Promise<void>;
  markAllRead(feedId: string): Promise<void>;
  refresh(): Promise<Feed[]>;
  /** Imports an OPML file (#97). */
  importOpml(text: string): Promise<ImportReport>;
  /** Address of the OPML export download. */
  opmlUrl(): string;
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

  /**
   * Imports an OPML file and then loads the new feeds (the import itself
   * only subscribes). Files over 2 MB are refused before uploading.
   */
  async importOpml(file: {
    size: number;
    text(): Promise<string>;
  }): Promise<{ report: ImportReport; feeds: Feed[] }> {
    if (file.size > MAX_OPML_BYTES) throw new Error(t('rss.opml_too_large'));
    const report = await this.gateway.importOpml(await file.text());
    const feeds = report.added > 0 ? await this.gateway.refresh() : await this.gateway.feeds();
    return { report, feeds };
  }

  opmlUrl(): string {
    return this.gateway.opmlUrl();
  }

  refresh(): Promise<Feed[]> {
    return this.gateway.refresh();
  }
}
