import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  NoteDTO,
  NoteRefDTO,
  RevisionDetailDTO,
  RevisionDTO,
  SearchHitDTO,
} from '../../../../shared/api-client/generated/notes';
import type { NoteInputDTO } from '../../../../shared/api-client/generated/notes-requests';
import type {
  DiffLine,
  Note,
  NoteInput,
  NoteRef,
  Revision,
  RevisionDetail,
  SearchHit,
} from '../../domain/note';
import type { NotesGateway } from '../../usecase/notesUseCases';

const toNote = (d: NoteDTO): Note => ({
  id: d.id,
  title: d.title,
  body: d.body,
  pinned: d.pinned,
  updatedAt: new Date(d.updatedAt),
  version: d.version ?? 0,
});

const path = (id: string) => `/api/notes/${encodeURIComponent(id)}`;
const toRevision = (d: RevisionDTO): Revision => ({
  id: d.id,
  title: d.title,
  createdAt: new Date(d.createdAt),
  size: d.size,
});

/** NotesGateway backed by the REST API. */
export class NotesApi implements NotesGateway {
  constructor(private readonly http: HttpClient) {}

  async list(pinnedOnly: boolean): Promise<Note[]> {
    return (await this.http.get<NoteDTO[]>(`/api/notes${pinnedOnly ? '?pinned=true' : ''}`)).map(toNote);
  }
  async save(
    input: NoteInput,
    id?: string,
    opts: { clientId?: string; version?: number } = {},
  ): Promise<Note> {
    // clientId (create) makes a replay idempotent; version (update) detects
    // changes made elsewhere meanwhile (#80).
    const body: NoteInputDTO = {
      ...input,
      ...(!id && opts.clientId ? { id: opts.clientId } : {}),
      ...(id && opts.version ? { version: opts.version } : {}),
    };
    const dto = id
      ? await this.http.put<NoteDTO>(`/api/notes/${encodeURIComponent(id)}`, body)
      : await this.http.post<NoteDTO>('/api/notes', body);
    return toNote(dto);
  }
  delete(id: string): Promise<void> {
    return this.http.delete(path(id));
  }
  async get(id: string): Promise<Note> {
    return toNote(await this.http.get<NoteDTO>(path(id)));
  }
  async search(query: string): Promise<SearchHit[]> {
    const hits = await this.http.get<SearchHitDTO[]>(`/api/notes/search?q=${encodeURIComponent(query)}`);
    return hits.map((h) => ({ id: h.id, title: h.title, snippet: h.snippet }));
  }
  async backlinks(id: string): Promise<NoteRef[]> {
    return (await this.http.get<NoteRefDTO[]>(`${path(id)}/backlinks`)).map((r) => ({
      id: r.id,
      title: r.title,
    }));
  }
  async revisions(id: string): Promise<Revision[]> {
    return (await this.http.get<RevisionDTO[]>(`${path(id)}/revisions`)).map(toRevision);
  }
  async revision(id: string, revisionId: string): Promise<RevisionDetail> {
    const d = await this.http.get<RevisionDetailDTO>(
      `${path(id)}/revisions/${encodeURIComponent(revisionId)}`,
    );
    return {
      ...toRevision(d),
      body: d.body,
      diff: d.diff.map((l) => ({ op: l.op as DiffLine['op'], text: l.text })),
    };
  }
  async restore(id: string, revisionId: string): Promise<Note> {
    return toNote(
      await this.http.post<NoteDTO>(`${path(id)}/revisions/${encodeURIComponent(revisionId)}/restore`),
    );
  }
}
