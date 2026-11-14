import { t } from '../../../shared/i18n';
import { type Article, type ArticleStatus, looksLikeUrl } from '../domain/readlater';

/** Output port: read-later backend. */
export interface ReadLaterGateway {
  list(status?: ArticleStatus): Promise<Article[]>;
  save(url: string): Promise<Article>;
  get(id: string): Promise<Article>;
  setStatus(id: string, status: ArticleStatus): Promise<void>;
  remove(id: string): Promise<void>;
}

/** Application logic of the read-later list (#99). */
export class ReadLaterUseCases {
  constructor(private readonly gateway: ReadLaterGateway) {}

  /** Unread and read articles, or those of one status. */
  list(status?: ArticleStatus): Promise<Article[]> {
    return this.gateway.list(status);
  }

  save(url: string): Promise<Article> {
    const u = url.trim();
    if (!looksLikeUrl(u)) return Promise.reject(new Error(t('readlater.url_invalid')));
    return this.gateway.save(u);
  }

  /** Opens an article; an unread one counts as read from now on. */
  async open(id: string): Promise<Article> {
    const a = await this.gateway.get(id);
    if (a.status === 'unread') {
      await this.gateway.setStatus(id, 'read');
      return { ...a, status: 'read' };
    }
    return a;
  }

  setStatus(id: string, status: ArticleStatus): Promise<void> {
    return this.gateway.setStatus(id, status);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}
