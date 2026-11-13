import { t } from '../../../shared/i18n';
import type { Bookmark, BookmarkInput, BookmarkSettings, Folder, Tag } from '../domain/bookmark';

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
  /** Only bookmarks whose last link check failed (#72). */
  broken?: boolean;
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
  importHtml(html: string): Promise<BookmarkImportResult>;
  getSettings(): Promise<BookmarkSettings>;
  saveSettings(s: BookmarkSettings): Promise<BookmarkSettings>;
  enrich(id: string): Promise<Bookmark>;
  checkLink(id: string): Promise<Bookmark>;
}

/** Result of a browser bookmark import. */
export interface BookmarkImportResult {
  created: number;
  skippedInvalid: number;
  skippedExists: number;
  foldersCreated: number;
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

  getSettings(): Promise<BookmarkSettings> {
    return this.gateway.getSettings();
  }

  saveSettings(s: BookmarkSettings): Promise<BookmarkSettings> {
    return this.gateway.saveSettings(s);
  }

  /** Fetches title (if none was entered) and favicon again (#72). */
  enrich(id: string): Promise<Bookmark> {
    return this.gateway.enrich(id);
  }

  /** Checks a link right away (#72). */
  checkLink(id: string): Promise<Bookmark> {
    return this.gateway.checkLink(id);
  }

  /** Imports a browser bookmark export (HTML). */
  importHtml(html: string): Promise<BookmarkImportResult> {
    if (!/<a\s/i.test(html)) return Promise.reject(new Error(t('bookmarks.no_bookmarks_in_file')));
    return this.gateway.importHtml(html);
  }
}
