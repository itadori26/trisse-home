import { ApiError } from '../../shared/api-client/http';
import {
  discard,
  emptyOutbox,
  enqueue,
  isPending,
  keepMine,
  type Entity,
  type OutboxState,
  type PendingChange,
  type ReplayResult,
  settle,
} from '../../shared/offline/domain/outbox';
import type { OutboxStore } from '../../shared/offline/adapter/outboxStore';

/** Sends one waiting change to the server (wired per entity). */
export type Sender = (c: PendingChange) => Promise<void>;
/** Reads the server's current version of an item after a 409 (0 = gone). */
export type VersionReader = (c: PendingChange) => Promise<number>;

/** Whether err means "no connection" rather than an answer of the server. */
export const isNetworkError = (err: unknown): boolean =>
  err instanceof TypeError || (err instanceof ApiError && err.status === 0);

/**
 * Offline sync of to-dos and notes (#80): keeps the outbox, replays it in
 * order when the connection is back and keeps refused changes as conflicts
 * for the user to decide. The feature use cases do not know about it; the
 * offline gateways in this folder record their changes here.
 */
export class OfflineSync {
  state = $state<OutboxState>(emptyOutbox());
  syncing = $state(false);
  /** Bumped after a replay changed server data, so widgets reload. */
  revision = $state(0);
  private readonly senders: Partial<Record<Entity, { send: Sender; version: VersionReader }>> = {};
  private loaded: Promise<void>;

  constructor(private readonly store: OutboxStore) {
    this.loaded = store.load().then((s) => {
      this.state = s;
    });
    if (typeof window !== 'undefined') window.addEventListener('online', () => void this.replay());
  }

  get pending(): number {
    return this.state.queue.length;
  }

  register(entity: Entity, send: Sender, version: VersionReader): void {
    this.senders[entity] = { send, version };
  }

  /** Waits for the persisted outbox (gateways overlay it on lists). */
  ready(): Promise<void> {
    return this.loaded;
  }

  isPending(entity: Entity, itemId: string): boolean {
    return isPending(this.state, entity, itemId);
  }

  async record(change: Omit<PendingChange, 'key' | 'at'>): Promise<void> {
    await this.loaded;
    this.update(enqueue(this.state, { ...change, key: crypto.randomUUID(), at: Date.now() }));
  }

  /** Sends the waiting changes in order; stops at the first network error. */
  async replay(): Promise<void> {
    await this.loaded;
    if (this.syncing || this.state.queue.length === 0) return;
    this.syncing = true;
    let changed = false;
    try {
      for (let first = this.state.queue[0]; first; first = this.state.queue[0]) {
        const result = await this.send(first);
        if (result.kind === 'offline') break;
        this.update(settle(this.state, result));
        changed = true;
      }
    } finally {
      this.syncing = false;
      if (changed) this.revision++;
    }
  }

  keepMine(key: string): void {
    this.update(keepMine(this.state, key));
    void this.replay();
  }

  discard(key: string): void {
    this.update(discard(this.state, key));
    this.revision++;
  }

  async clear(): Promise<void> {
    this.state = emptyOutbox();
    await this.store.clear();
  }

  private async send(c: PendingChange): Promise<ReplayResult> {
    const handler = this.senders[c.entity];
    if (!handler) return { kind: 'offline' };
    try {
      await handler.send(c);
      return { kind: 'ok' };
    } catch (err) {
      if (isNetworkError(err)) return { kind: 'offline' };
      if (err instanceof ApiError && err.status === 404 && c.op !== 'create') return { kind: 'gone' };
      if (err instanceof ApiError && err.status === 409) {
        const serverVersion = await handler.version(c).catch(() => 0);
        return serverVersion ? { kind: 'changed', serverVersion } : { kind: 'gone' };
      }
      return { kind: 'rejected', message: err instanceof Error ? err.message : String(err) };
    }
  }

  private update(next: OutboxState): void {
    this.state = next;
    void this.store.save(next);
  }
}
