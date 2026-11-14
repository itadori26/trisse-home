import { emptyOutbox, type OutboxState } from '../domain/outbox';

/** Persists the outbox (#80); survives reloads and closing the app. */
export interface OutboxStore {
  load(): Promise<OutboxState>;
  save(state: OutboxState): Promise<void>;
  clear(): Promise<void>;
}

const DB = 'trisse-offline';
const STORE = 'outbox';
const KEY = 'state';

function open(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB, 1);
    req.onupgradeneeded = () => req.result.createObjectStore(STORE);
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error ?? new Error('IndexedDB unavailable'));
  });
}

function run<T>(mode: IDBTransactionMode, fn: (s: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  return open().then(
    (db) =>
      new Promise<T>((resolve, reject) => {
        const tx = db.transaction(STORE, mode);
        const req = fn(tx.objectStore(STORE));
        tx.oncomplete = () => (db.close(), resolve(req.result));
        tx.onerror = () => (db.close(), reject(tx.error ?? new Error('IndexedDB failed')));
      }),
  );
}

/** IndexedDB store; falls back to memory where IndexedDB is not available (private windows). */
export class IndexedDbOutboxStore implements OutboxStore {
  private memory: OutboxState = emptyOutbox();
  private readonly available = typeof indexedDB !== 'undefined';

  async load(): Promise<OutboxState> {
    if (!this.available) return this.memory;
    try {
      const v = (await run('readonly', (s) => s.get(KEY))) as OutboxState | undefined;
      return v && Array.isArray(v.queue) && Array.isArray(v.conflicts) ? v : emptyOutbox();
    } catch {
      return this.memory;
    }
  }

  async save(state: OutboxState): Promise<void> {
    this.memory = state;
    if (!this.available) return;
    await run('readwrite', (s) => s.put(plain(state), KEY)).catch(() => undefined);
  }

  async clear(): Promise<void> {
    this.memory = emptyOutbox();
    if (!this.available) return;
    await run('readwrite', (s) => s.delete(KEY)).catch(() => undefined);
  }
}

/** Structured clone needs plain objects (no Svelte proxies). */
function plain(state: OutboxState): OutboxState {
  return JSON.parse(JSON.stringify(state)) as OutboxState;
}

/** Removes the outbox database (logout). */
export async function deleteOutbox(): Promise<void> {
  if (typeof indexedDB === 'undefined') return;
  await new Promise<void>((resolve) => {
    const req = indexedDB.deleteDatabase(DB);
    req.onsuccess = req.onerror = req.onblocked = () => resolve();
  });
}
