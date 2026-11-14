import { describe, expect, it } from 'vitest';
import { ApiError } from '../../shared/api-client/http';
import { emptyOutbox, type OutboxState } from '../../shared/offline/domain/outbox';
import type { OutboxStore } from '../../shared/offline/adapter/outboxStore';
import type { Todo } from '../../features/todos/domain/todo';
import type { TodosGateway } from '../../features/todos/usecase/todosUseCases';
import { OfflineTodosGateway } from './offlineTodos';
import { OfflineSync } from './sync.svelte';

class MemoryStore implements OutboxStore {
  state: OutboxState = emptyOutbox();
  async load() {
    return this.state;
  }
  async save(s: OutboxState) {
    this.state = s;
  }
  async clear() {
    this.state = emptyOutbox();
  }
}

const todo = (id: string, title: string, version: number, done = false): Todo => ({
  id,
  title,
  due: '',
  done,
  overdue: false,
  repeat: { kind: '', every: 1, days: [] },
  reminder: { at: '', daysBefore: 0 },
  version,
});

/** A server with to-dos and version checks; offline = every call fails like fetch does. */
class FakeServer implements TodosGateway {
  offline = false;
  items = new Map<string, Todo>();
  calls: string[] = [];
  private net() {
    if (this.offline) throw new TypeError('Failed to fetch');
  }
  async list() {
    this.net();
    return [...this.items.values()];
  }
  async create(title: string, due: string, id?: string) {
    this.net();
    this.calls.push(`create ${id}`);
    const existing = id ? this.items.get(id) : undefined;
    if (existing) return existing;
    const t = { ...todo(id ?? 'srv', title, 1), due };
    this.items.set(t.id, t);
    return t;
  }
  async update(id: string, title: string, due: string, done: boolean, _s?: unknown, version?: number) {
    this.net();
    this.calls.push(`update ${id} v${version ?? 0}`);
    const cur = this.items.get(id);
    if (!cur) throw new ApiError('not found', 404, 'not_found');
    if (version && version !== cur.version) throw new ApiError('changed', 409, 'conflict');
    const t = { ...cur, title, due, done, version: cur.version + 1 };
    this.items.set(id, t);
    return t;
  }
  async delete(id: string) {
    this.net();
    this.items.delete(id);
  }
  async clearCompleted() {
    return 0;
  }
}

describe('offline to-dos (#80)', () => {
  it('records changes offline, shows them and replays them in order', async () => {
    const server = new FakeServer();
    server.items.set('a', todo('a', 'Milch', 1));
    const sync = new OfflineSync(new MemoryStore());
    const gw = new OfflineTodosGateway(server, sync);
    expect((await gw.list(false)).map((t) => t.title)).toEqual(['Milch']);

    server.offline = true;
    const created = await gw.create('Brot', '');
    await gw.update('a', 'Milch kaufen', '', true);
    expect(sync.pending).toBe(2);
    // The cached list shows what was changed offline.
    expect((await gw.list(true)).map((t) => `${t.title}:${t.done}`)).toEqual([
      'Milch kaufen:true',
      'Brot:false',
    ]);
    expect((await gw.list(false)).map((t) => t.title)).toEqual(['Brot']);

    // Still offline: replay stops and keeps everything.
    await sync.replay();
    expect(sync.pending).toBe(2);

    server.offline = false;
    await sync.replay();
    expect(sync.pending).toBe(0);
    expect(server.calls).toEqual([`create ${created.id}`, 'update a v1']);
    expect(server.items.get('a')).toMatchObject({ title: 'Milch kaufen', done: true, version: 2 });
    expect(sync.revision).toBe(1);
  });

  it('turns a change made elsewhere into a conflict the user decides', async () => {
    const server = new FakeServer();
    server.items.set('a', todo('a', 'Milch', 1));
    const sync = new OfflineSync(new MemoryStore());
    const gw = new OfflineTodosGateway(server, sync);
    await gw.list(true);
    server.offline = true;
    await gw.update('a', 'Hafermilch', '', false);
    // Meanwhile someone else edits it on the server.
    server.items.set('a', todo('a', 'Milch (2 l)', 2));
    server.offline = false;
    await sync.replay();
    expect(sync.state.conflicts).toEqual([expect.objectContaining({ kind: 'changed', serverVersion: 2 })]);
    expect(server.items.get('a')?.title).toBe('Milch (2 l)');

    sync.keepMine(sync.state.conflicts[0]?.change.key ?? '');
    await sync.replay();
    expect(sync.state.conflicts).toHaveLength(0);
    expect(server.items.get('a')).toMatchObject({ title: 'Hafermilch', version: 3 });
  });

  it('does not queue changes the server refused while online', async () => {
    const server = new FakeServer();
    const sync = new OfflineSync(new MemoryStore());
    const gw = new OfflineTodosGateway(server, sync);
    await expect(gw.update('missing', 'x', '', false)).rejects.toBeInstanceOf(ApiError);
    expect(sync.pending).toBe(0);
  });
});
