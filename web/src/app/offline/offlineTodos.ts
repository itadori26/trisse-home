import type { Reminder, Repeat, Todo } from '../../features/todos/domain/todo';
import type { TodosGateway } from '../../features/todos/usecase/todosUseCases';
import { overlay, type PendingChange } from '../../shared/offline/domain/outbox';
import { isNetworkError, type OfflineSync } from './sync.svelte';

type Schedule = { repeat: Repeat; reminder: Reminder };

const NO_REPEAT: Repeat = { kind: '', every: 1, days: [] };
const NO_REMINDER: Reminder = { at: '', daysBefore: 0 };

function fromChange(c: PendingChange): Todo {
  const p = c.payload as { title?: string; due?: string; done?: boolean; schedule?: Schedule };
  return {
    id: c.itemId,
    title: p.title ?? '',
    due: p.due ?? '',
    done: p.done ?? false,
    overdue: false,
    repeat: p.schedule?.repeat ?? NO_REPEAT,
    reminder: p.schedule?.reminder ?? NO_REMINDER,
    version: 0,
  };
}

/**
 * TodosGateway that keeps working offline (#80): changes that fail for lack
 * of a connection go to the outbox and show up in lists right away; the
 * versions of the last loaded list go along so the server can detect
 * changes made elsewhere meanwhile.
 */
export class OfflineTodosGateway implements TodosGateway {
  private readonly versions = new Map<string, number>();
  private last: Todo[] = [];

  constructor(
    private readonly inner: TodosGateway,
    private readonly sync: OfflineSync,
  ) {
    sync.register(
      'todo',
      (c) => this.send(c),
      (c) => this.currentVersion(c.itemId),
    );
  }

  async list(includeDone: boolean): Promise<Todo[]> {
    await this.sync.ready();
    let items: Todo[];
    try {
      items = await this.inner.list(includeDone);
      this.last = items;
      for (const t of items) this.versions.set(t.id, t.version);
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      items = this.last;
    }
    const merged = overlay(items, this.sync.state, 'todo', fromChange);
    return includeDone ? merged : merged.filter((t) => !t.done);
  }

  async create(title: string, due: string): Promise<Todo> {
    const id = crypto.randomUUID();
    try {
      return await this.inner.create(title, due, id);
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      await this.sync.record({
        entity: 'todo',
        op: 'create',
        itemId: id,
        payload: { title, due, done: false },
        baseVersion: 0,
      });
      return fromChange({
        key: '',
        entity: 'todo',
        op: 'create',
        itemId: id,
        payload: { title, due },
        baseVersion: 0,
        at: 0,
      });
    }
  }

  async update(id: string, title: string, due: string, done: boolean, schedule?: Schedule): Promise<Todo> {
    const version = this.versions.get(id) ?? 0;
    try {
      const t = await this.inner.update(id, title, due, done, schedule, version);
      this.versions.set(id, t.version);
      return t;
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      const payload = { title, due, done, ...(schedule ? { schedule } : {}) };
      await this.sync.record({ entity: 'todo', op: 'update', itemId: id, payload, baseVersion: version });
      return {
        ...fromChange({
          key: '',
          entity: 'todo',
          op: 'update',
          itemId: id,
          payload,
          baseVersion: version,
          at: 0,
        }),
        version,
      };
    }
  }

  async delete(id: string): Promise<void> {
    try {
      await this.inner.delete(id);
    } catch (err) {
      if (!isNetworkError(err)) throw err;
      await this.sync.record({
        entity: 'todo',
        op: 'delete',
        itemId: id,
        payload: {},
        baseVersion: this.versions.get(id) ?? 0,
      });
    }
  }

  clearCompleted(): Promise<number> {
    return this.inner.clearCompleted();
  }

  /** Replays one waiting change (create, update or delete). */
  private async send(c: PendingChange): Promise<void> {
    const p = c.payload as { title: string; due: string; done: boolean; schedule?: Schedule };
    if (c.op === 'delete') return this.inner.delete(c.itemId);
    if (c.op === 'create') {
      const t = await this.inner.create(p.title, p.due, c.itemId);
      if (p.done || p.schedule)
        await this.inner.update(c.itemId, p.title, p.due, p.done, p.schedule, t.version);
      return;
    }
    const t = await this.inner.update(c.itemId, p.title, p.due, p.done, p.schedule, c.baseVersion);
    this.versions.set(c.itemId, t.version);
  }

  private async currentVersion(id: string): Promise<number> {
    const t = (await this.inner.list(true)).find((x) => x.id === id);
    if (t) this.versions.set(id, t.version);
    return t?.version ?? 0;
  }
}
