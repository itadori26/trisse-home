import type { HttpClient } from '../../../../shared/api-client/http';
import type { ArticleDTO } from '../../../../shared/api-client/generated/readlater';
import type { Article, ArticleStatus } from '../../domain/readlater';
import type { ReadLaterGateway } from '../../usecase/readLaterUseCases';

const toArticle = (d: ArticleDTO): Article => ({
  ...d,
  content: d.content ?? '',
  savedAt: new Date(d.savedAt),
});

/** ReadLaterGateway backed by the REST API. */
export class ReadLaterApi implements ReadLaterGateway {
  constructor(private readonly http: HttpClient) {}

  async list(status?: ArticleStatus): Promise<Article[]> {
    return (await this.http.get<ArticleDTO[]>(`/api/readlater${status ? `?status=${status}` : ''}`)).map(
      toArticle,
    );
  }
  async save(url: string): Promise<Article> {
    return toArticle(await this.http.post<ArticleDTO>('/api/readlater', { url }));
  }
  async get(id: string): Promise<Article> {
    return toArticle(await this.http.get<ArticleDTO>(`/api/readlater/${encodeURIComponent(id)}`));
  }
  setStatus(id: string, status: ArticleStatus): Promise<void> {
    return this.http.put(`/api/readlater/${encodeURIComponent(id)}/status`, { status });
  }
  remove(id: string): Promise<void> {
    return this.http.delete(`/api/readlater/${encodeURIComponent(id)}`);
  }
}
