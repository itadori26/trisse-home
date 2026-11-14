/**
 * Offline outbox (#80): changes to to-dos and notes made without a
 * connection wait here and are replayed in order once the server is back.
 * Pure functions; persistence and replay live in adapters.
 */

export type Entity = 'todo' | 'note';
export type Op = 'create' | 'update' | 'delete';

/** One waiting change. */
export interface PendingChange {
  /** ID of the change itself (stable across reloads). */
  key: string;
  entity: Entity;
  op: Op;
  /** The item; for creates a client-generated ID the server keeps. */
  itemId: string;
  /** The fields to send (create/update). */
  payload: Record<string, unknown>;
  /** Version the change was made from (update/delete); 0 = no check. */
  baseVersion: number;
  /** When it was made (ms). */
  at: number;
}

/** A change the server refused: changed elsewhere, or rejected. */
export interface Conflict {
  change: PendingChange;
  kind: 'changed' | 'rejected';
  /** The server's current version (changed) or its message (rejected). */
  serverVersion: number;
  message: string;
}

export interface OutboxState {
  queue: PendingChange[];
  conflicts: Conflict[];
}

export const emptyOutbox = (): OutboxState => ({ queue: [], conflicts: [] });

/**
 * Adds a change, merged with what already waits for the same item so the
 * server sees as few requests as possible:
 * - create + update → one create with the newer fields
 * - update + update → one update from the first change's version
 * - create + delete → nothing (the server never saw the item)
 * - update + delete → one delete from the first change's version
 */
export function enqueue(state: OutboxState, change: PendingChange): OutboxState {
  const queue = [...state.queue];
  const i = queue.findIndex((c) => c.entity === change.entity && c.itemId === change.itemId);
  if (i < 0) return { ...state, queue: [...queue, change] };
  const prev = queue[i] as PendingChange;
  if (change.op === 'update' && prev.op !== 'delete') {
    queue[i] = { ...prev, payload: { ...prev.payload, ...change.payload }, at: change.at };
  } else if (change.op === 'delete' && prev.op === 'create') {
    queue.splice(i, 1);
  } else if (change.op === 'delete') {
    queue[i] = { ...change, baseVersion: prev.baseVersion };
  } else {
    queue.push(change); // e.g. a create after a delete of the same ID: keep both, in order
  }
  return { ...state, queue };
}

/** The result of sending one change. */
export type ReplayResult =
  | { kind: 'ok' }
  | { kind: 'gone' } // update/delete of an item that no longer exists
  | { kind: 'changed'; serverVersion: number }
  | { kind: 'rejected'; message: string }
  | { kind: 'offline' };

/** Applies the result of sending the first change; offline keeps the queue as it is. */
export function settle(state: OutboxState, result: ReplayResult): OutboxState {
  const [first, ...rest] = state.queue;
  if (!first || result.kind === 'offline') return state;
  switch (result.kind) {
    case 'ok':
    case 'gone':
      return { ...state, queue: rest };
    case 'changed':
      return {
        queue: rest,
        conflicts: [
          ...state.conflicts,
          { change: first, kind: 'changed', serverVersion: result.serverVersion, message: '' },
        ],
      };
    case 'rejected':
      return {
        queue: rest,
        conflicts: [
          ...state.conflicts,
          { change: first, kind: 'rejected', serverVersion: 0, message: result.message },
        ],
      };
  }
}

/** "Keep mine": send the change again on top of the server's version. */
export function keepMine(state: OutboxState, key: string): OutboxState {
  const c = state.conflicts.find((x) => x.change.key === key);
  if (!c || c.kind !== 'changed') return state;
  return {
    conflicts: state.conflicts.filter((x) => x !== c),
    queue: [...state.queue, { ...c.change, baseVersion: c.serverVersion }],
  };
}

/** "Keep server" (or discard a rejected change): drop it. */
export function discard(state: OutboxState, key: string): OutboxState {
  return { ...state, conflicts: state.conflicts.filter((x) => x.change.key !== key) };
}

/** Whether an item has a change waiting (for the "not synced yet" marker). */
export const isPending = (state: OutboxState, entity: Entity, itemId: string): boolean =>
  state.queue.some((c) => c.entity === entity && c.itemId === itemId);

/**
 * Shows the waiting changes on top of a (possibly cached) list, so what was
 * changed offline does not disappear on the next load: creates are added,
 * updates applied, deletes removed.
 */
export function overlay<T extends { id: string }>(
  items: T[],
  state: OutboxState,
  entity: Entity,
  build: (c: PendingChange) => T,
): T[] {
  let out = [...items];
  for (const c of state.queue.filter((x) => x.entity === entity)) {
    const i = out.findIndex((x) => x.id === c.itemId);
    if (c.op === 'delete') {
      if (i >= 0) out.splice(i, 1);
    } else if (i >= 0) {
      out[i] = { ...out[i], ...(c.payload as Partial<T>) } as T;
    } else if (c.op === 'create') {
      out = [...out, build(c)];
    }
  }
  return out;
}
