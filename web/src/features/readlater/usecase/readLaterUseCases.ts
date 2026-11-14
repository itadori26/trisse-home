import { t } from '../../../shared/i18n';
import { type Article, type ArticleStatus, type ImportReport, looksLikeUrl } from '../domain/readlater';

/** Output port: read-later backend. */
export interface ReadLaterGateway {
  list(status?: ArticleStatus): Promise<Article[]>;
  save(url: string): Promise<Article>;
  get(id: string): Promise<Article>;
  setStatus(id: string, status: ArticleStatus): Promise<void>;
  remove(id: string): Promise<void>;
  archiveDays(): Promise<number>;
  saveArchiveDays(days: number): Promise<number>;
  importPocket(csv: string): Promise<ImportReport>;
}

/** Application logic of the read-later list (#99). */
export class ReadLaterUseCases {
  private status: Promise<boolean> | null = null;

  constructor(private readonly gateway: ReadLaterGateway) {}

  /** Whether saving is possible here (module on, not a guest); asked once. */
  available(): Promise<boolean> {
    this.status ??= this.gateway.archiveDays().then(
      () => true,
      () => false,
    );
    return this.status;
  }

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

  /** Days after which archived articles are deleted (0 = keep). */
  archiveDays(): Promise<number> {
    return this.gateway.archiveDays();
  }
  saveArchiveDays(days: number): Promise<number> {
    return this.gateway.saveArchiveDays(days);
  }
  /** Imports a Pocket CSV export; the reader copies follow in the background. */
  async importPocket(file: { size: number; text(): Promise<string> }): Promise<ImportReport> {
    if (file.size > 5 << 20) throw new Error(t('readlater.import_too_large'));
    return this.gateway.importPocket(await file.text());
  }
}
