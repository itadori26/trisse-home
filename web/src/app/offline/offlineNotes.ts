import type {
  Note,
  NoteInput,
  NoteRef,
  Revision,
  RevisionDetail,
  SearchHit,
} from '../../features/notes/domain/note';
import type { NotesGateway } from '../../features/notes/usecase/notesUseCases';
import { overlay, type PendingChange } from '../../shared/offline/domain/outbox';
import { isNetworkError, type OfflineSync } from './sync.svelte';

function fromChange(c: PendingChange): Note {
  const p = c.payload as Partial<NoteInput>;
  return {
    id: c.itemId,
    title: p.title ?? '',
    body: p.body ?? '',
    pinned: p.pinned ?? false,
    updatedAt: new Date(c.at || Date.now()),
    version: 0,
  };
}

/**
 * NotesGateway that keeps working offline (#80) for creating, editing and
 * deleting; search, history and backlinks need the server.
 */
export class OfflineNotesGateway implements NotesGateway {
  private readonly versions = new Map<string, number>();
  private last: Note[] = [];

  constructor(
    private readonly inner: NotesGateway,
    private readonly sync: OfflineSync,
  ) {
    sync.register(
      'note',
      (c) => this.send(c),
      (c) => this.currentVersion(c.itemId),
    );
  }

  async list(pinnedOnly: boolean): Promise<Note[]> {
    await this.sync.ready();
    let items: Note[];
    try {
      items = await this.inner.list(pinnedOnly);
      if (!pinnedOnly) this.last = items;
      for (const n of items) this.versions.set(n.id, n.version);
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      items = pinnedOnly ? this.last.filter((n) => n.pinned) : this.last;
    }
    const merged = overlay(items, this.sync.state, 'note', fromChange);
    return pinnedOnly ? merged.filter((n) => n.pinned) : merged;
  }

  async save(input: NoteInput, id?: string): Promise<Note> {
    if (!id) {
      const clientId = crypto.randomUUID();
      try {
        return await this.inner.save(input, undefined, { clientId });
      } catch (err) {
        if (!isNetworkError(err)) throw err;
        const change = {
          entity: 'note' as const,
          op: 'create' as const,
          itemId: clientId,
          payload: { ...input },
          baseVersion: 0,
        };
        await this.sync.record(change);
        return fromChange({ ...change, key: '', at: Date.now() });
      }
    }
    const version = this.versions.get(id) ?? 0;
    try {
      const n = await this.inner.save(input, id, { version });
      this.versions.set(id, n.version);
      return n;
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      const change = {
        entity: 'note' as const,
        op: 'update' as const,
        itemId: id,
        payload: { ...input },
        baseVersion: version,
      };
      await this.sync.record(change);
      return { ...fromChange({ ...change, key: '', at: Date.now() }), version };
    }
  }

  async delete(id: string): Promise<void> {
    try {
      await this.inner.delete(id);
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      await this.sync.record({
        entity: 'note',
        op: 'delete',
        itemId: id,
        payload: {},
        baseVersion: this.versions.get(id) ?? 0,
      });
    }
  }

  get(id: string): Promise<Note> {
    return this.inner.get(id);
  }
  search(query: string): Promise<SearchHit[]> {
    return this.inner.search(query);
  }
  backlinks(id: string): Promise<NoteRef[]> {
    return this.inner.backlinks(id);
  }
  revisions(id: string): Promise<Revision[]> {
    return this.inner.revisions(id);
  }
  revision(id: string, revisionId: string): Promise<RevisionDetail> {
    return this.inner.revision(id, revisionId);
  }
  restore(id: string, revisionId: string): Promise<Note> {
    return this.inner.restore(id, revisionId);
  }

  private async send(c: PendingChange): Promise<void> {
    const input = c.payload as unknown as NoteInput;
    if (c.op === 'delete') return this.inner.delete(c.itemId);
    if (c.op === 'create') {
      await this.inner.save(input, undefined, { clientId: c.itemId });
      return;
    }
    const n = await this.inner.save(input, c.itemId, { version: c.baseVersion });
    this.versions.set(c.itemId, n.version);
  }

  private async currentVersion(id: string): Promise<number> {
    const n = await this.inner.get(id);
    this.versions.set(id, n.version);
    return n.version;
  }
}
