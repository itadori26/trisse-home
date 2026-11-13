import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  BookmarkDTO,
  FolderDTO,
  ListDTO,
  SettingsDTO,
} from '../../../../shared/api-client/generated/bookmarks';
import type {
  BookmarkInputDTO,
  FolderInputDTO,
  SettingsInputDTO,
} from '../../../../shared/api-client/generated/bookmarks-requests';
import type { Bookmark, BookmarkInput, BookmarkSettings, Folder, LinkStatus } from '../../domain/bookmark';
import type {
  BookmarkFilter,
  BookmarkImportResult,
  BookmarkList,
  BookmarksGateway,
} from '../../usecase/bookmarksUseCases';

const toBookmark = (d: BookmarkDTO): Bookmark => ({
  id: d.id,
  folderId: d.folderId,
  title: d.title,
  url: d.url,
  description: d.description,
  tags: d.tags,
  link: {
    state: (d.link?.state ?? '') as LinkStatus['state'],
    code: d.link?.code ?? 0,
    target: d.link?.target ?? '',
    checkedAt: d.link?.checkedAt ?? '',
  },
  hasIcon: d.hasIcon ?? false,
});
const toFolder = (d: FolderDTO): Folder => ({ id: d.id, name: d.name, count: d.count });
const toDTO = (i: BookmarkInput): BookmarkInputDTO => ({
  folderId: i.folderId,
  title: i.title,
  url: i.url,
  description: i.description,
  tags: i.tags,
});
const path = (id: string) => `/api/bookmarks/${encodeURIComponent(id)}`;
const folderPath = (id: string) => `/api/bookmarks/folders/${encodeURIComponent(id)}`;

/** BookmarksGateway backed by the REST API. */
export class BookmarksApi implements BookmarksGateway {
  constructor(private readonly http: HttpClient) {}

  async list(filter: BookmarkFilter): Promise<BookmarkList> {
    const q = new URLSearchParams();
    if (filter.folderId) q.set('folder', filter.folderId);
    if (filter.tag) q.set('tag', filter.tag);
    if (filter.query) q.set('q', filter.query);
    if (filter.broken) q.set('status', 'broken');
    const dto = await this.http.get<ListDTO>(`/api/bookmarks?${q}`);
    return { bookmarks: dto.bookmarks.map(toBookmark), folders: dto.folders.map(toFolder), tags: dto.tags };
  }
  async create(input: BookmarkInput): Promise<Bookmark> {
    return toBookmark(await this.http.post<BookmarkDTO>('/api/bookmarks', toDTO(input)));
  }
  async update(id: string, input: BookmarkInput): Promise<Bookmark> {
    return toBookmark(await this.http.put<BookmarkDTO>(path(id), toDTO(input)));
  }
  delete(id: string): Promise<void> {
    return this.http.delete(path(id));
  }
  async createFolder(name: string): Promise<Folder> {
    const body: FolderInputDTO = { name };
    return toFolder(await this.http.post<FolderDTO>('/api/bookmarks/folders', body));
  }
  async renameFolder(id: string, name: string): Promise<Folder> {
    const body: FolderInputDTO = { name };
    return toFolder(await this.http.patch<FolderDTO>(folderPath(id), body));
  }
  deleteFolder(id: string): Promise<void> {
    return this.http.delete(folderPath(id));
  }

  async getSettings(): Promise<BookmarkSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/bookmarks/settings')) };
  }
  async saveSettings(s: BookmarkSettings): Promise<BookmarkSettings> {
    const body: SettingsInputDTO = { enrich: s.enrich, checkLinks: s.checkLinks };
    return { ...(await this.http.put<SettingsDTO>('/api/bookmarks/settings', body)) };
  }
  async enrich(id: string): Promise<Bookmark> {
    return toBookmark(await this.http.post<BookmarkDTO>(`${path(id)}/enrich`));
  }
  async checkLink(id: string): Promise<Bookmark> {
    return toBookmark(await this.http.post<BookmarkDTO>(`${path(id)}/check`));
  }

  importHtml(html: string): Promise<BookmarkImportResult> {
    return this.http.postText<BookmarkImportResult>(
      '/api/bookmarks/import',
      html,
      'text/html; charset=utf-8',
    );
  }
}
