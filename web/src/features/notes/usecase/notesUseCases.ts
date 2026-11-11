import { t } from '../../../shared/i18n';
import type { Note, NoteInput, NotesScope } from '../domain/note';

/** Output port: notes backend. */
export interface NotesGateway {
  list(pinnedOnly: boolean): Promise<Note[]>;
  save(input: NoteInput, id?: string): Promise<Note>;
  delete(id: string): Promise<void>;
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
}
