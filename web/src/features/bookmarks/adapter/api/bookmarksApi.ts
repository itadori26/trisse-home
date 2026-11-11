import type { HttpClient } from '../../../../shared/api-client/http';
import type { BookmarkDTO, FolderDTO, ListDTO } from '../../../../shared/api-client/generated/bookmarks';
import type {
  BookmarkInputDTO,
  FolderInputDTO,
} from '../../../../shared/api-client/generated/bookmarks-requests';
import type { Bookmark, BookmarkInput, Folder } from '../../domain/bookmark';
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
});
const toFolder = (d: FolderDTO): Folder => ({ id: d.id, name: d.name, count: d.count });
const toDTO = (i: BookmarkInput): BookmarkInputDTO => ({ ...i });
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

  importHtml(html: string): Promise<BookmarkImportResult> {
    return this.http.postText<BookmarkImportResult>(
      '/api/bookmarks/import',
      html,
      'text/html; charset=utf-8',
    );
  }
}
