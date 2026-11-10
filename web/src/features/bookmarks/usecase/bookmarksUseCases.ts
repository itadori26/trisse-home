import type { Bookmark, BookmarkInput, Folder, Tag } from '../domain/bookmark';

/** Result of listing bookmarks. */
export interface BookmarkList {
  bookmarks: Bookmark[];
  folders: Folder[];
  tags: Tag[];
}

/** Filter for listing bookmarks. */
export interface BookmarkFilter {
  folderId?: string;
  tag?: string;
  query?: string;
}

/** Output port: bookmarks backend. */
export interface BookmarksGateway {
  list(filter: BookmarkFilter): Promise<BookmarkList>;
  create(input: BookmarkInput): Promise<Bookmark>;
  update(id: string, input: BookmarkInput): Promise<Bookmark>;
  delete(id: string): Promise<void>;
  createFolder(name: string): Promise<Folder>;
  renameFolder(id: string, name: string): Promise<Folder>;
  deleteFolder(id: string): Promise<void>;
}

/** Application logic of the bookmarks feature. */
export class BookmarksUseCases {
  constructor(private readonly gateway: BookmarksGateway) {}

  list(filter: BookmarkFilter = {}): Promise<BookmarkList> {
    return this.gateway.list(filter);
  }

  /** Creates or updates a bookmark; URLs without scheme get https://. */
  save(input: BookmarkInput, id?: string): Promise<Bookmark> {
    const url = input.url.trim();
    const normalised = { ...input, url: /^[a-z][a-z0-9+.-]*:/i.test(url) ? url : `https://${url}` };
    return id ? this.gateway.update(id, normalised) : this.gateway.create(normalised);
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  createFolder(name: string): Promise<Folder> {
    return this.gateway.createFolder(name.trim());
  }

  renameFolder(id: string, name: string): Promise<Folder> {
    return this.gateway.renameFolder(id, name.trim());
  }

  deleteFolder(id: string): Promise<void> {
    return this.gateway.deleteFolder(id);
  }
}
