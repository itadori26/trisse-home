import { t } from '../../../shared/i18n';
import type {
  Note,
  NoteInput,
  NoteRef,
  NotesScope,
  Revision,
  RevisionDetail,
  SearchHit,
} from '../domain/note';

/** Output port: notes backend. */
export interface NotesGateway {
  list(pinnedOnly: boolean): Promise<Note[]>;
  save(input: NoteInput, id?: string, opts?: { clientId?: string; version?: number }): Promise<Note>;
  delete(id: string): Promise<void>;
  get(id: string): Promise<Note>;
  search(query: string): Promise<SearchHit[]>;
  backlinks(id: string): Promise<NoteRef[]>;
  revisions(id: string): Promise<Revision[]>;
  revision(id: string, revisionId: string): Promise<RevisionDetail>;
  restore(id: string, revisionId: string): Promise<Note>;
}

/** Error for input rejected before it reaches the backend. */
export class NoteValidationError extends Error {
  override name = 'NoteValidationError';
}

/** Application logic of the notes feature. */
export class NotesUseCases {
  constructor(private readonly gateway: NotesGateway) {}

  list(scope: NotesScope): Promise<Note[]> {
    return this.gateway.list(scope === 'pinned');
  }

  /** Saves a note; it needs a title or some text. */
  save(input: NoteInput, id?: string): Promise<Note> {
    if (input.title.trim() === '' && input.body.trim() === '') {
      return Promise.reject(new NoteValidationError(t('notes.empty_note')));
    }
    return this.gateway.save(input, id);
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  get(id: string): Promise<Note> {
    return this.gateway.get(id);
  }

  /** Full-text search (#73); queries shorter than two characters find nothing. */
  async search(query: string): Promise<SearchHit[]> {
    return query.trim().length < 2 ? [] : this.gateway.search(query.trim());
  }

  /** Notes that link here with [[Title]] (#73). */
  backlinks(id: string): Promise<NoteRef[]> {
    return this.gateway.backlinks(id);
  }

  /** Earlier versions, newest first (#73). */
  revisions(id: string): Promise<Revision[]> {
    return this.gateway.revisions(id);
  }

  revision(id: string, revisionId: string): Promise<RevisionDetail> {
    return this.gateway.revision(id, revisionId);
  }

  /** Makes an earlier version current; the replaced text is kept as a version. */
  restore(id: string, revisionId: string): Promise<Note> {
    return this.gateway.restore(id, revisionId);
  }
}
