import type { HttpClient } from '../../../../shared/api-client/http';
import type { NoteDTO } from '../../../../shared/api-client/generated/notes';
import type { NoteInputDTO } from '../../../../shared/api-client/generated/notes-requests';
import type { Note, NoteInput } from '../../domain/note';
import type { NotesGateway } from '../../usecase/notesUseCases';

const toNote = (d: NoteDTO): Note => ({
  id: d.id,
  title: d.title,
  body: d.body,
  pinned: d.pinned,
  updatedAt: new Date(d.updatedAt),
});

/** NotesGateway backed by the REST API. */
export class NotesApi implements NotesGateway {
  constructor(private readonly http: HttpClient) {}

  async list(pinnedOnly: boolean): Promise<Note[]> {
    return (await this.http.get<NoteDTO[]>(`/api/notes${pinnedOnly ? '?pinned=true' : ''}`)).map(toNote);
  }
  async save(input: NoteInput, id?: string): Promise<Note> {
    const body: NoteInputDTO = { ...input };
    const dto = id
      ? await this.http.put<NoteDTO>(`/api/notes/${encodeURIComponent(id)}`, body)
      : await this.http.post<NoteDTO>('/api/notes', body);
    return toNote(dto);
  }
  delete(id: string): Promise<void> {
    return this.http.delete(`/api/notes/${encodeURIComponent(id)}`);
  }
}
